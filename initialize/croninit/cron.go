// /initialize/croninit/cron.go
package croninit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var (
	c = cron.New()
)

const (
	// 心跳文件路径
	heartbeatFile = "/tmp/data_collect_heartbeat"
	// 心跳间隔（秒）
	heartbeatInterval = "*/10 * * * * *" // 每10秒执行一次
)

// 定义任务初始化
func CronInit() {

	// 添加心跳任务
	c.AddFunc(heartbeatInterval, func() {
		if err := sendHeartbeat(); err != nil {
			fmt.Printf("发送心跳包失败: %v\n", err)
		}
	})

	c.Start()
}

func sendHeartbeat() error {
	// 确保目录存在
	dir := filepath.Dir(heartbeatFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}

	// 创建或更新心跳文件
	file, err := os.OpenFile(heartbeatFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logrus.Errorf("打开心跳文件失败: %v", err)
		return err
	}
	defer file.Close()

	// 更新文件时间戳
	now := time.Now()
	if err := os.Chtimes(heartbeatFile, now, now); err != nil {
		logrus.Errorf("更新心跳文件时间戳失败: %v", err)
		return err
	}
	return nil
}
