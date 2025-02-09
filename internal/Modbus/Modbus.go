package modbus

import (
	"dataCollect/mqtt/publish"
	"encoding/json"
	"fmt"
	"log"
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
	WindSpeed     float32 `json:"wind_speed,omitempty"`
	WindDirection int     `json:"wind_direction,omitempty"`
	Humidity      float32 `json:"humidity,omitempty"`
	Temperature   float32 `json:"temperature,omitempty"`
}

var modbusClient modbus.Client

// 定义要读取的寄存器列表
var registers = []Register{
	{"风速", 500, 1, handlerWindSpeed}, // 读取地址 0x0000 开始的 2 个寄存器
	// {"风向", 503, 1, handlerWindDirection},
	// {"湿度", 504, 1, handlerHumidity},
	// {"温度", 505, 1, handlerTemperature},
}
var macAddr = "0F0F0F0F0F0F"
var cfgID = "0e7ee1c0-41f3-a9fc-4897-b938d5895d9d"

func ModbusInit() error {
	// 创建 Modbus RTU 客户端
	handler := modbus.NewRTUClientHandler("/dev/ttyS1")
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1

	handler.SlaveId = 1
	handler.Timeout = time.Second * 1

	// 连接 Modbus
	err := handler.Connect()
	if err != nil {
		logrus.Error("Modbus 连接失败: %v", err)
		return err
	}
	defer handler.Close()

	modbusClient = modbus.NewClient(handler)
	// 创建气象站设备，可以重复发送，因为服务器端有去重判断
	RegisterDev()

	// 进入数据读取循环
	go ModbusLoop()
	return nil
}
func ModbusLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	for {
		select {
		case <-ticker.C:
			readData()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

type RegisterSt struct {
	CfgID string `json:"cfgID"`
	Mac   string `json:"mac"`
}

func RegisterDev() {
	topic := "devices/register"
	var dev RegisterSt
	dev.CfgID = cfgID
	dev.Mac = macAddr
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
		logrus.Printf("读取 %s (地址: 0x%04X, 长度: %d)...\n", reg.Name, reg.Address, reg.Length)
		results, err := modbusClient.ReadHoldingRegisters(reg.Address, reg.Length)
		if err != nil {
			log.Printf("读取 %s 失败: %v\n", reg.Name, err)
			continue
		}

		// 解析结果
		for i := 0; i < len(results); i += 2 {
			value, err := reg.Handler(results)
			if err != nil {
				logrus.Printf("解析 %s 失败: %v\n", reg.Name, err)
				continue
			}
			fileVale[reg.Name] = value
			logrus.Printf("  %s: %.2f\n", reg.Name, value)
		}
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
		}
	}
	payload, err := json.Marshal(WeatherData)
	if err != nil {
		logrus.Printf("json Marshal err:%v\n", err)

	}
	publish.PublishMessage(genTopic(), payload)
}

// 处理温度  TODO 处理温度为负的情况
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
	return float32(value) / 100, nil // 温度值放大了 10 倍，需要除以 10
}

func genTopic() string {
	topic := "devices/telemetry"
	return fmt.Sprintf("%s/%s/%s", topic, cfgID, macAddr)
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
