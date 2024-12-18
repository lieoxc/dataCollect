package mqtt

import (
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var MqttConfig Config

type Config struct {
	Broker            string          `json:"broker"`
	User              string          `json:"user"`
	Pass              string          `json:"pass"`
	ChannelBufferSize int             `json:"channel_buffer_size"`
	WriteWorkers      int             `json:"write_workers"`
	Telemetry         Telemetry       `json:"telemetry"`
	Attributes        AttributeConfig `json:"attributes"`
	Commands          TopicConfig     `json:"commands"`
	Events            TopicConfig     `json:"events"`
	OTA               OTATopicConfig  `json:"ota"`
}
type Telemetry struct {
	SubscribeTopic        string `json:"subscribe_topic"`
	PublishTopic          string `json:"publish_topic"`
	GatewaySubscribeTopic string `json:"gateway_subscribe_topic"`
	GatewayPublishTopic   string `json:"gateway_publish_topic"`
	QoS                   int    `json:"qos"`
	PoolSize              int    `json:"pool_size"`
	BatchSize             int    `json:"batch_size"`
}

type TopicConfig struct {
	PublishTopic          string `json:"publish_topic"`
	SubscribeTopic        string `json:"subscribe_topic"`
	GatewaySubscribeTopic string `json:"gateway_subscribe_topic"`
	GatewayPublishTopic   string `json:"gateway_publish_topic"`
	QoS                   int    `json:"qos"`
}

type AttributeConfig struct {
	SubscribeTopic                string `json:"subscribe_topic"`
	GatewaySubscribeTopic         string `json:"gateway_subscribe_topic"`
	PublishResponseTopic          string `json:"publish_response_topic"`
	GatewayPublishResponseTopic   string `json:"gateway_publish_response_topic"`
	PublishTopic                  string `json:"publish_topic"`
	GatewayPublishTopic           string `json:"gateway_publish_topic"`
	SubscribeResponseTopic        string `json:"subscribe_response_topic"`
	GatewaySubscribeResponseTopic string `json:"gateway_subscribe_response_topic"`
	PublishGetTopic               string `json:"publish_get_topic"`
	GatewayPublishGetTopic        string `json:"gateway_publish_get_topic"`
	QoS                           int    `json:"qos"`
}

type OTATopicConfig struct {
	PublishTopic   string `json:"publish_topic"`
	SubscribeTopic string `json:"subscribe_topic"`
	QoS            int    `json:"qos"`
}

func MqttInit() error {
	// 初始化配置
	err := loadConfig()
	if err != nil {
		return err
	}
	return nil
}
func loadConfig() error {
	var configMap map[string]interface{}
	// 将 map 映射到 MQTTConfig 结构体
	// 注意！！！yml文件中带_的key，是无法通过UnmarshalKey解析的
	err := viper.Unmarshal(&configMap)
	if err != nil {
		return fmt.Errorf("unable to decode into struct, %s", err)
	}
	// 将 map 转换为 json
	jsonStr, err := json.Marshal(configMap["mqtt"])
	if err != nil {
		return fmt.Errorf("unable to marshal config, %s", err)
	}
	// 将 json 转换为结构体
	err = json.Unmarshal(jsonStr, &MqttConfig)
	if err != nil {
		return fmt.Errorf("unable to unmarshal config, %s", err)
	}

	// 打印配置
	logrus.Debug("mqtt config:", MqttConfig)

	// 单独获取 broker 配置
	broker := viper.GetString("mqtt.broker")
	if broker == "" {
		broker = "localhost:1883"
		logrus.Println("Using default broker:", broker)
	}
	MqttConfig.Broker = broker

	// 单独获取 user 配置
	user := viper.GetString("mqtt.user")
	if user == "" {
		user = "root"
		logrus.Println("Using default user:", user)
	}
	MqttConfig.User = user

	// 单独获取 pass 配置
	pass := viper.GetString("mqtt.pass")
	if pass == "" {
		pass = "root"
		logrus.Println("Using default pass:", pass)
	}
	MqttConfig.Pass = pass

	// 单独获取 channel_buffer_size 配置
	channelBufferSize := viper.GetInt("mqtt.channel_buffer_size")
	if channelBufferSize == 0 {
		channelBufferSize = 10000
		logrus.Println("Using default channel_buffer_size:", channelBufferSize)
	}
	MqttConfig.ChannelBufferSize = channelBufferSize

	// 单独获取 write_workers 配置
	writeWorkers := viper.GetInt("mqtt.write_workers")
	if writeWorkers == 0 {
		writeWorkers = 10
		logrus.Println("Using default write_workers:", writeWorkers)
	}
	MqttConfig.WriteWorkers = writeWorkers

	// 单独获取 pool_size 配置
	poolSize := viper.GetInt("mqtt.telemetry.pool_size")
	if poolSize == 0 {
		poolSize = 100
		logrus.Println("Using default pool_size:", poolSize)
	}
	MqttConfig.Telemetry.PoolSize = poolSize

	// 单独获取 batch_size 配置
	batchSize := viper.GetInt("mqtt.telemetry.batch_size")
	if batchSize == 0 {
		batchSize = 100
		logrus.Println("Using default batch_size:", batchSize)
	}
	return nil
}
