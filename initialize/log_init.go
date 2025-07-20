package initialize

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

type customFormatter struct {
	logrus.TextFormatter
}

func (*customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var levelText string
	switch entry.Level {
	case logrus.DebugLevel:
		levelText = "DEBUG" // 蓝色
	case logrus.InfoLevel:
		levelText = "INFO " // 青色
	case logrus.WarnLevel:
		levelText = "WARN " // 黄色
	case logrus.ErrorLevel:
		levelText = "ERROR" // 红色
	case logrus.FatalLevel, logrus.PanicLevel:
		levelText = "FATAL" // 红色，更重要的错误
	default:
		levelText = "UNKNOWN" // 默认颜色
	}

	// 获取调用者信息
	var fileAndLine string
	if entry.HasCaller() {
		dir := filepath.Dir(entry.Caller.File)
		fileAndLine = fmt.Sprintf("%s/%s:%d", filepath.Base(dir), filepath.Base(entry.Caller.File), entry.Caller.Line)
	}

	// 组装格式化字符串
	msg := fmt.Sprintf("%s [%s] [%s] %s\n",
		levelText, // 日志级别，带颜色
		entry.Time.Format("2006-01-02 15:04:05.9999"), // 时间戳，下划线加颜色
		fileAndLine,   // 文件名:行号，带颜色
		entry.Message, // 日志消息
	)
	return []byte(msg), nil
}
func LogInIt() {

	// 初始化 Logrus,不创建logrus实例，直接使用包级别的函数，这样可以在项目的任何地方使用logrus，目前不考虑多日志模块的情况
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&customFormatter{logrus.TextFormatter{
		//ForceColors:   true,
		FullTimestamp: true,
	}})

	var logLevels = map[string]logrus.Level{
		"panic": logrus.PanicLevel,
		"fatal": logrus.FatalLevel,
		"error": logrus.ErrorLevel,
		"warn":  logrus.WarnLevel,
		"info":  logrus.InfoLevel,
		"debug": logrus.DebugLevel,
		"trace": logrus.TraceLevel,
	}

	levelStr := viper.GetString("log.level")
	if level, ok := logLevels[levelStr]; ok {
		logrus.SetLevel(level)
	} else {
		logrus.Error("Invalid log level in config, setting to default level")
		logrus.SetLevel(logrus.InfoLevel) // 设置默认级别
	}
	// 设置日志输出到文件
	currentTime := time.Now().Format("2006-01-02-1504") // 格式化为年月日时分
	path := viper.GetString("log.path")
	var logPath string
	if path == "" {
		logPath = filepath.Join("logs", currentTime+".log")
	} else {
		logPath = filepath.Join(path, currentTime+".log")
	}

	logrus.SetOutput(&lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    20,      // 每个日志文件的最大大小，单位为MB
		MaxBackups: 5,       // 保留旧日志文件的最大数量
		MaxAge:     30,      // 保留旧日志文件的最大天数
		Compress:   true,    // 是否压缩旧日志文件
	})

	logrus.Debug("*************************** dataCollect Init Finsh**********************")
}
