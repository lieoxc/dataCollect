package subscribe

import (
	config "dataCollect/mqtt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/go-basic/uuid"
	"github.com/panjf2000/ants/v2"
	"github.com/sirupsen/logrus"
)

var SubscribeMqttClient mqtt.Client
var TelemetryMessagesChan chan map[string]interface{}

func GenTopic(topic string) string {
	//topic = path.Join("$share/mygroup", topic)
	return topic
}
func SubscribeInit() error {
	// 创建mqtt客户端
	subscribeMqttClient()
	// 创建消息队列
	telemetryMessagesChan()

	//消息订阅
	err := subscribe()
	return err
}

// 创建消息队列
func telemetryMessagesChan() {
	TelemetryMessagesChan = make(chan map[string]interface{}, config.MqttConfig.ChannelBufferSize)
	writeWorkers := config.MqttConfig.WriteWorkers
	for i := 0; i < writeWorkers; i++ {
		go MessagesChanHandler(TelemetryMessagesChan)
	}
}

func subscribe() error {
	//订阅telemetry消息
	err := SubscribeTelemetry()
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}
func subscribeMqttClient() {
	// 初始化配置
	opts := mqtt.NewClientOptions()
	opts.AddBroker(config.MqttConfig.Broker)
	opts.SetUsername(config.MqttConfig.User)
	opts.SetPassword(config.MqttConfig.Pass)
	id := "dataCollect-" + uuid.New()[0:8]
	opts.SetClientID(id)
	logrus.Info("clientid: ", id)

	// 干净会话
	opts.SetCleanSession(true)
	// 恢复客户端订阅，需要broker支持
	opts.SetResumeSubs(true)
	// 自动重连
	opts.SetAutoReconnect(true)
	opts.SetConnectRetryInterval(5 * time.Second)
	opts.SetMaxReconnectInterval(200 * time.Second)
	// 消息顺序
	opts.SetOrderMatters(false)
	opts.SetOnConnectHandler(func(_ mqtt.Client) {
		logrus.Println("mqtt connect success")
	})
	// 断线重连
	opts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		logrus.Println("mqtt connect  lost: ", err)
		SubscribeMqttClient.Disconnect(250)
		for {
			if token := SubscribeMqttClient.Connect(); token.Wait() && token.Error() != nil {
				logrus.Error("MQTT Broker 1 连接失败:", token.Error())
				time.Sleep(5 * time.Second)
				continue
			}
			subscribe()
			break
		}
	})

	SubscribeMqttClient = mqtt.NewClient(opts)
	// 等待连接成功，失败重新连接
	for {
		if token := SubscribeMqttClient.Connect(); token.Wait() && token.Error() != nil {
			logrus.Error("MQTT Broker 1 连接失败:", token.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}
}

// 订阅telemetry消息
func SubscribeTelemetry() error {
	//如果配置了别的数据库，遥测数据不写入原来的库了
	p, err := ants.NewPool(config.MqttConfig.Telemetry.PoolSize)
	if err != nil {
		return err
	}
	deviceTelemetryMessageHandler := func(_ mqtt.Client, d mqtt.Message) {
		err = p.Submit(func() {
			// 处理消息
			TelemetryMessages(d.Payload(), d.Topic())
		})
		if err != nil {
			logrus.Error(err)
		}
	}

	topic := config.MqttConfig.Telemetry.SubscribeTopic
	topic = GenTopic(topic)
	logrus.Info("subscribe topic:", topic)

	qos := byte(config.MqttConfig.Telemetry.QoS)

	if token := SubscribeMqttClient.Subscribe(topic, qos, deviceTelemetryMessageHandler); token.Wait() && token.Error() != nil {
		logrus.Error(token.Error())
		return err
	}
	return nil
}

func SubscribeOtaUpprogress() error {
	// 订阅ota升级消息
	otaUpgradeHandler := func(_ mqtt.Client, d mqtt.Message) {
		// 处理消息
		logrus.Debug("ota upgrade message:", string(d.Payload()))
		OtaUpgrade(d.Payload(), d.Topic())
	}
	topic := config.MqttConfig.OTA.SubscribeTopic
	qos := byte(config.MqttConfig.OTA.QoS)
	if token := SubscribeMqttClient.Subscribe(topic, qos, otaUpgradeHandler); token.Wait() && token.Error() != nil {
		logrus.Error(token.Error())
		return token.Error()
	}
	return nil
}
