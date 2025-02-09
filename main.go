package main

import (
	"dataCollect/initialize"
	modbus "dataCollect/internal/Modbus"
	mqttapp "dataCollect/mqtt"
	"dataCollect/mqtt/publish"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
)

func main() {
	initialize.ViperInit("./configs/conf.yml")
	initialize.LogInIt()

	err := mqttapp.MqttInit()
	if err != nil {
		logrus.Fatal(err)
	}
	publish.CreateMqttClient()
	if err != nil {
		logrus.Fatal(err)
	}
	modbus.ModbusInit()

	gracefulShutdown()
}

func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Println("dataCollect exiting")
}
