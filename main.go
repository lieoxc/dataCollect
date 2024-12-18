package main

import (
	"dataCollect/initialize"
	mqttapp "dataCollect/mqtt"
	"dataCollect/mqtt/subscribe"
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
	err = subscribe.SubscribeInit()
	if err != nil {
		logrus.Fatal(err)
	}

	gracefulShutdown()

}

func gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	logrus.Println("dataCollect exiting")
}
