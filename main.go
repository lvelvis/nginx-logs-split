package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	//日志目录
	srcDirPath := "/usr/local/nginx/logs"

	//存放切割日志目录
	targetDirPath := "/usr/local/nginx/logs/history"

	//ngixn进程ID文件
	nginxPidPath := "/usr/local/nginx/logs/nginx.pid"

	//检查存放切割日志目录是否存在，如果不存在则创建
	finfo, errFile := os.Stat(targetDirPath)
	if errFile != nil {
		errFile := os.MkdirAll(targetDirPath, 0777)
		if errFile != nil {
			fmt.Println("创建日志目录失败：" + errFile.Error())
			return
		}
	} else if !finfo.IsDir() {
		fmt.Println(targetDirPath + "已经存在且不是一个目录")
		return
	}

	//获取当前日期，作为此次切割日志根目录
	t := time.Now()
	nowDateTime := t.Format("2006-01-02")
	logPath := targetDirPath + "/" + nowDateTime
	os.MkdirAll(logPath, 0777)

	//获取nginx的进程ID
	pfile, err := os.Open(nginxPidPath)
	defer pfile.Close()
	if err != nil {
		fmt.Println("not found nginx pid file")
		return
	}
	pidData, _ := ioutil.ReadAll(pfile)
	pid := string(pidData)
	pid = strings.Replace(pid, "\n", "", -1)

	//遍历日志目录
	filepath.Walk(srcDirPath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		} else {
			//获取切割日志路径
			targetfilePath := strings.Replace(path, srcDirPath, logPath, 1)
			if strings.Index(targetfilePath, "nginx.pid") != -1 {
				return nil
			}

			//移动文件
			syscall.Rename(path, targetfilePath)

			//创建原文件,这里不需要了，因为重启nginx后会自动生成滴
			// nFile,errCreate := os.Create(path)
			// if errCreate != nil {
			// 	fmt.Println("create file faild:"+errCreate.Error())
			// }
			// defer nFile.Close()
		}
		return nil
	})

	//平滑重启nginx
	cmd := exec.Command("kill", "-USR1", pid)
	_, errCmd := cmd.Output()
	if errCmd != nil {
		fmt.Println("重启nginx失败：" + errCmd.Error())
		return
	}
	fmt.Println("success")
}
