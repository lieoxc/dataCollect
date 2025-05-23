package modbus

import (
	"context"
	"dataCollect/initialize"
	"dataCollect/mqtt/publish"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/goburrow/modbus"
	"github.com/sirupsen/logrus"
)

// 寄存器信息结构体
type Register struct {
	Name    string                            // 寄存器名称
	Address uint16                            // 起始地址
	Length  uint16                            // 读取的寄存器数量
	Handler func([]byte) (interface{}, error) // 处理读取数据的函数
}

// 定义气象站设备监控的字段
type WeatherSt struct {
	WindSpeed      float32 `json:"wind_speed"`
	WindDirection  int     `json:"wind_direction"`
	Humidity       float32 `json:"humidity"`
	Temperature    float32 `json:"temperature"`
	Rainfall       float32 `json:"rainfall"`
	SolarRadiation float32 `json:"solarRadiation"`
}

type AtributeSt struct {
	Latitude     string `json:"latitude"`
	Longitude    string `json:"longitude"`
	Signal       string `json:"signal"`
	ModelVersion string `json:"modelVersion"`
	Network      string `json:"network"`
}

var modbusClient modbus.Client

// 定义要读取的寄存器列表
var registers = []Register{
	{"风速", 500, 1, handlerWindSpeed}, // 读取地址 0x0000 开始的 2 个寄存器
	{"风向", 503, 1, handlerWindDirection},
	{"湿度", 504, 1, handlerHumidity},
	{"温度", 505, 1, handlerTemperature},
	{"雨量", 513, 1, handlerRainfall},
	{"太阳辐射", 515, 1, handlerSolarRadiation},
}

var MacAddr = "0F0F0F0F0F0F"                       // 气象监控站的设备ID针对每个路由器都是唯一的
var cfgID = "964d6220-ecbf-a043-1960-85b1a2758cea" // 气象监控站的模板ID

func ModbusInit() error {
	// 创建 Modbus RTU 客户端
	handler := modbus.NewRTUClientHandler("/dev/ttyS1")
	handler.BaudRate = 4800
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1

	handler.SlaveId = 1
	handler.Timeout = time.Second * 1

	// 连接 Modbus
	err := handler.Connect()
	if err != nil {
		logrus.Errorf("Modbus 连接失败: %v", err)
		return err
	}
	defer handler.Close()

	modbusClient = modbus.NewClient(handler)
	// 创建气象站设备，可以重复发送，因为服务器端有去重判断
	RegisterDev()

	// 进入数据读取循环
	go ModbusLoop()
	go attributesLoop()
	return nil
}
func attributesLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	for {
		select {
		case <-ticker.C:
			ReportAttributes()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}
func ReportAttributes() {
	var reportSt AtributeSt
	// 1. 获取GPS
	dataGps, err := initialize.Redis.HGetAll(context.Background(), "gps_data").Result()
	if err != nil {
		logrus.Error("gps_data err:%v", err)
	}
	if val, ok := dataGps["latitude"]; ok {
		reportSt.Latitude = val
	}
	if val, ok := dataGps["longitude"]; ok {
		reportSt.Longitude = val
	}
	// 2. 获取4G
	dataModem, err := initialize.Redis.HGetAll(context.Background(), "modem_data").Result()
	if err != nil {
		logrus.Error("modem_data err:%v", err)
	}
	if val, ok := dataModem["signal"]; ok {
		reportSt.Signal = val
	}
	if val, ok := dataModem["modelVersion"]; ok {
		reportSt.ModelVersion = val
	}
	if val, ok := dataModem["network"]; ok {
		reportSt.Network = val
	}
	payload, err := json.Marshal(reportSt)
	if err != nil {
		logrus.Debugf("json Marshal err:%v\n", err)
		return
	}
	publish.PublishMessage(genAttributesTopic(), payload)
}
func resetRainfall() error {
	// 使用功能码 0x06 (写单个寄存器)
	// 地址 6002H (24578 十进制)
	// 写入值 0x5A (10 十进制)
	modbusClient.WriteSingleRegister(24578, 90)
	logrus.Debug("resetRainfall Send Succeed.")
	return nil
}
func ModbusLoop() {
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	count := 0
	for {
		select {
		case <-ticker.C:
			readData()
			count++
			//定时30min 发送雨量清0
			if count == 180 { // 10 * 180 = 1800s = 30min
				if err := resetRainfall(); err != nil {
					logrus.Error("雨量清零失败: ", err)
				}
				count = 0 // 重置计数器
			}
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

type RegisterSt struct {
	CfgID string `json:"cfgID"`
	Mac   string `json:"mac"`
	Name  string `json:"name"`
}

func RegisterDev() {
	topic := "devices/register"
	var dev RegisterSt
	dev.CfgID = cfgID
	addr, err := getMACAddress("eth0")
	if err != nil {
		logrus.Errorf("getMACAddresserr:%v ", err)
	}
	if addr != "" {
		MacAddr = addr
	}
	dev.Mac = MacAddr
	dev.Name = "气象监控站"
	payload, err := json.Marshal(dev)
	if err != nil {
		logrus.Printf("json Marshal err:%v\n", err)

	}
	publish.PublishMessage(topic, payload)
}
func readData() {
	fileVale := make(map[string]interface{})
	// 读取每个寄存器并输出结果
	for _, reg := range registers {
		//logrus.Printf("读取 %s (地址: 0x%04X, 长度: %d)...\n", reg.Name, reg.Address, reg.Length)
		results, err := modbusClient.ReadHoldingRegisters(reg.Address, reg.Length)
		if err != nil {
			logrus.Errorf("读取 %s 失败: %v", reg.Name, err)
			continue
		}

		// 解析结果
		for i := 0; i < len(results); i += 2 {
			value, err := reg.Handler(results)
			if err != nil {
				logrus.Warnf("解析 %s 失败: %v", reg.Name, err)
				continue
			}
			fileVale[reg.Name] = value
			logrus.Debugf("  %s: %v", reg.Name, value)
		}
	}
	if len(fileVale) == 0 {
		logrus.Warn("can not read any data from modbus")
		return
	}
	var WeatherData WeatherSt
	for name, value := range fileVale {
		switch name {
		case "风速":
			WeatherData.WindSpeed, _ = value.(float32)
		case "风向":
			WeatherData.WindDirection, _ = value.(int)
		case "湿度":
			WeatherData.Humidity, _ = value.(float32)
		case "温度":
			WeatherData.Temperature, _ = value.(float32)
		case "雨量":
			WeatherData.Rainfall, _ = value.(float32)
		case "太阳辐射":
			WeatherData.SolarRadiation, _ = value.(float32)
		}
	}
	payload, err := json.Marshal(WeatherData)
	if err != nil {
		logrus.Debugf("json Marshal err:%v\n", err)
		return
	}
	publish.PublishMessage(genTopic(), payload)
}

// 处理雨量
func handlerRainfall(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return float32(value) / 10, nil // 雨量值放大了 10 倍，需要除以 10
}

// 处理太阳辐射
func handlerSolarRadiation(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return float32(value), nil // 太阳辐射返回的就是实际值
}

// 处理温度
func handlerTemperature(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return float32(value) / 10, nil // 温度值放大了 10 倍，需要除以 10
}

// 处理湿度
func handlerHumidity(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return float32(value) / 10, nil // 温度值放大了 10 倍，需要除以 10
}

// 处理风向
func handlerWindDirection(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return int(value), nil // 温度值放大了 10 倍，需要除以 10
}

// 处理风速
func handlerWindSpeed(data []byte) (interface{}, error) {
	// 假设读取的是 16 位有符号整数，转换为温度值
	if len(data) < 2 {
		return nil, fmt.Errorf("数据长度不足")
	}
	value := int16(data[0])<<8 | int16(data[1])
	return float32(value) / 10, nil // 风速放大了 10 倍，需要除以 10
}

func genTopic() string {
	topic := "devices/telemetry"
	return fmt.Sprintf("%s/%s/%s", topic, cfgID, MacAddr)
}
func genAttributesTopic() string {
	topic := "devices/attributes"
	return fmt.Sprintf("%s/%s/%s", topic, cfgID, MacAddr)
}
func getMACAddress(interfaceName string) (string, error) {
	//return "1C:40:E8:11:69:54", nil
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", err
	}
	macClean := strings.ReplaceAll(strings.ToUpper(iface.HardwareAddr.String()), ":", "")
	return macClean, nil
}
