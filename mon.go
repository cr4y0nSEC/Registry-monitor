package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/sys/windows/registry"
)

var (
	registryPathToPoll string
	valueName          string
	pollInterval       time.Duration
	logFilePath        string
)

func init1() {
	flag.StringVar(&registryPathToPoll, "path", "", "The registry path to poll.")
	flag.StringVar(&valueName, "value", "", "The name of the value to check.")
	flag.DurationVar(&pollInterval, "interval", 5*time.Second, "The polling interval.")
	flag.StringVar(&logFilePath, "log", "registry_poll.log", "The log file path.")
	flag.Parse()

	if registryPathToPoll == "" || valueName == "" {
		log.Fatal("Both -path and -value flags are required.")
	}
}

func pollRegistryForChanges(keyPath string, valueName string, logFilePath string, interval time.Duration) {
	var lastValue string
	var lastCheckTime time.Time

	// 确保日志文件目录存在

	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开日志文件：%v", err)
	}
	defer logFile.Close()

	for {
		k, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE)
		if err != nil {
			log.Printf("打开注册表键时出错：%v，将在下次轮询时重试", err)
			time.Sleep(interval)
			continue
		}
		defer k.Close()

		value, _, err := k.GetStringValue(valueName)
		if err != nil {
			log.Printf("读取注册表值时出错：%v，将在下次轮询时重试", err)
			time.Sleep(interval)
			continue
		}

		if lastValue != "" && lastValue != value {
			elapsed := time.Since(lastCheckTime)
			logEntry := fmt.Sprintf("注册表键 %s 下的值 %s 已更改为 %s，距离上次检查已过去 %v\n", keyPath, lastValue, value, elapsed)
			if _, err := logFile.WriteString(logEntry); err != nil {
				log.Fatalf("无法向日志文件写入内容：%v", err)
			}
			fmt.Println(logEntry)
		}

		lastValue = value
		lastCheckTime = time.Now()
		time.Sleep(interval)
	}
}

func main() {
	// 初始化和命令行参数解析

	init1()

	fmt.Printf("开始轮询注册表键的更改：%s\n", registryPathToPoll)
	fmt.Println("按 'c' 键退出程序...")

	pollRegistryForChanges(registryPathToPoll, valueName, logFilePath, pollInterval)
	var input string
	fmt.Scanln(&input)

	if input == "c" || input == "C" {
		fmt.Println("用户按下 'c' 键，退出程序")
		os.Exit(0)
	}

}
