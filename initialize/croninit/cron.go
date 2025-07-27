// /initialize/croninit/cron.go
package croninit

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
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

	c.AddFunc("0 */60 * * * *", func() {
		logrus.Debug("Log File Clean:")
		CleanupLogs()
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

func CleanupLogs() error {
	logDir := "/mnt/data_collect/logs"
	cutoffTime := time.Now().AddDate(0, 0, -3) // 3天前的时间

	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue // 跳过子目录
		}

		filename := file.Name()
		// 检查文件扩展名
		if filepath.Ext(filename) != ".log" {
			continue
		}

		// 提取文件名中的时间部分（不含扩展名）
		timePart := strings.TrimSuffix(filename, ".log")
		if len(timePart) < 14 { // 确保至少包含日期和时间部分
			continue
		}

		// 解析文件名中的时间
		fileTime, err := time.Parse("2006-01-02-1504", timePart)
		if err != nil {
			continue // 跳过格式不匹配的文件
		}

		// 检查文件时间是否超过3天
		if fileTime.Before(cutoffTime) {
			filePath := filepath.Join(logDir, filename)
			err := os.Remove(filePath)
			if err != nil {
				logrus.Errorf("CleanupLogs %s failed.", filePath)
			}
		}
	}
	return nil
}
