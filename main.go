package main

import (
	"dataCollect/initialize"
	"dataCollect/initialize/croninit"
	modbus "dataCollect/internal/Modbus"
	mqttapp "dataCollect/mqtt"
	"dataCollect/mqtt/publish"
	"flag"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

func main() {
	// 1. 定义命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "./configs/conf.yml", "Path to config file")
	flag.Parse()

	initialize.ViperInit(configPath)
	initialize.LogInIt()
	croninit.CronInit()
	err := mqttapp.MqttInit()
	if err != nil {
		logrus.Fatal(err)
	}
	publish.CreateMqttClient()
	if err != nil {
		logrus.Fatal(err)
	}
	initialize.RedisInit()
	modbus.ModbusInit()

	gracefulShutdown()
}

func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Println("dataCollect exiting")
}
