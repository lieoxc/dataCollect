package initialize

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

var Redis *redis.Client

func RedisInit() error {
	conf, err := loadConfig()
	if err != nil {
		return fmt.Errorf("加载redis配置失败:%v", err)
	}
	client := connectRedis(conf)
	if checkRedisClient(client) != nil {
		return fmt.Errorf("连接redis失败:%v", err)
	}

	Redis = client
	return nil
}

func connectRedis(conf *RedisConfig) *redis.Client {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})
	// 如果返回nil，就创建这个DB

	return redisClient
}
func checkRedisClient(redisClient *redis.Client) error {
	// 通过 cient.Ping() 来检查是否成功连接到了 redis 服务器
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return err
	} else {
		logrus.Debug("连接redis成完成...")
		return nil
	}
}

func loadConfig() (*RedisConfig, error) {
	redisConfig := &RedisConfig{
		Addr:     viper.GetString("db.redis.addr"),
		Password: viper.GetString("db.redis.password"),
		DB:       viper.GetInt("db.redis.db"),
	}

	if redisConfig.Addr == "" {
		redisConfig.Addr = "localhost:6379"
	}
	return redisConfig, nil
}
