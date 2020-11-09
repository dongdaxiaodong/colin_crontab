package main

import (
	"crontab/master"
	"flag"
	"fmt"
	"runtime"
	"time"
)

var(
	confFile string//配置文件路径
)

//解析命令行参数
func initArgs(){
	//master -config ./master.json
	//master -h
	flag.StringVar(&confFile,"config","./master.json","传入master.json")
	flag.Parse()
}

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var(
		err error
	)
	//初始化线程
	initEnv()

	//加载配置
	if err = master.InitConfig("./master.json"); err != nil {
		goto ERR
	}

	//任务管理器
	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}
	//启动api http服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	//正常退出
	for{
		time.Sleep(1 * time.Second)
	}
	return

ERR:
	fmt.Println(err)
}