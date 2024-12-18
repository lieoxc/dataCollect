package subscribe

import (
	"github.com/sirupsen/logrus"
)

// 对列处理，数据入库
func MessagesChanHandler(messages <-chan map[string]interface{}) {
	logrus.Println("批量写入协程启动")
	// var telemetryList []*model.TelemetryData

	// batchSize := config.MqttConfig.Telemetry.BatchSize
	// logrus.Println("每次最大写入条数：", batchSize)
	// for {
	// 	for i := 0; i < batchSize; i++ {
	// 		// 获取消息
	// 		//logrus.Debug("管道消息数量:", len(messages))
	// 		message, ok := <-messages
	// 		if !ok {
	// 			break
	// 		}

	// 		if tskv, ok := message["telemetryData"].(model.TelemetryData); ok {
	// 			telemetryList = append(telemetryList, &tskv)
	// 		} else {
	// 			logrus.Error("管道消息格式错误")
	// 		}
	// 		// 如果管道没有消息，则检查入库
	// 		if len(messages) > 0 {
	// 			continue
	// 		}
	// 		break
	// 	}

	// 	// 如果tskvList有数据，则写入数据库
	// 	if len(telemetryList) > 0 {
	// 		logrus.Info("批量写入遥测数据表的条数:", len(telemetryList))
	// 		err := dal.CreateTelemetrDataBatch(telemetryList)
	// 		if err != nil {
	// 			logrus.Error(err)
	// 		}

	// 		// 更新当前值表
	// 		err = dal.UpdateTelemetrDataBatch(telemetryList)
	// 		if err != nil {
	// 			logrus.Error(err)
	// 		}

	// 		// 清空telemetryList
	// 		telemetryList = []*model.TelemetryData{}
	// 	}
	// }
}

// 处理消息
func TelemetryMessages(payload []byte, topic string) {
	//如果配置了别的数据库，遥测数据不写入原来的库了
	// dbType := viper.GetString("grpc.tptodb_type")
	// if dbType == "TSDB" || dbType == "KINGBASE" || dbType == "POLARDB" {
	// 	logrus.Infof("do not insert db for dbType:%v", dbType)
	// 	return
	// }

	logrus.Debugln(string(payload))
	// 验证消息有效性
	telemetryPayload, err := verifyPayload(payload)
	if err != nil {
		logrus.Error(err.Error(), topic)
		return
	}
	// device, err := initialize.GetDeviceCacheById(telemetryPayload.DeviceId)
	// if err != nil {
	// 	logrus.Error(err.Error())
	// 	return
	// }
	TelemetryMessagesHandle(telemetryPayload.Values, topic)
}

func TelemetryMessagesHandle(telemetryBody []byte, topic string) {
	// TODO脚本处理

}
