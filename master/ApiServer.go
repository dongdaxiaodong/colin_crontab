package master

import (
	"crontab/common"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiServer struct {
	httpServer *http.Server
}

var (
	//单例对象
	G_apiServer *ApiServer
)

//保存任务接口
//POST job = {"name":"job1","command":"echo hello","cronExpr"}
func handleJobSave(resp http.ResponseWriter,req *http.Request){
	//任务保存在ETCD中
	//1.解析POST表单
	var(
		err error
		postJob string
		job common.Job
		oldJob *common.Job
		bytes []byte
	)
	//1.解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	//2.取表单中的job字段
	postJob = req.PostForm.Get("job")
	//3,.反序列化job
	if err = json.Unmarshal([]byte(postJob),&job);err != nil {
		goto ERR
	}

	//保存到etcd
	if oldJob,err = G_jobMgr.SaveJob(&job);err != nil {
		goto ERR
	}


	//5返回正常应答,({errno:0,msg:"","data":{}})
	if bytes,err = common.BuildResponse(0,"success",oldJob); err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes,err = common.BuildResponse(-1,err.Error(),nil); err == nil {
		resp.Write(bytes)
	}
}


//删除任务接口
//job/delte
func handleJobDelete(resp http.ResponseWriter,req *http.Request){
	var (
		err error
		name string
		oldJob *common.Job
		bytes []byte
	)
	//post: a=1&b=2&c=3
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//删除的任务名
	name = req.PostForm.Get("name")
	//删除任务
	if oldJob,err = G_jobMgr.DeleteJob(name);err != nil {
		goto ERR
	}
	//正常应答
	if bytes,err = common.BuildResponse(0,"success",oldJob);err == nil {
		resp.Write(bytes)
	}

	return

ERR:
	fmt.Println("hello")
	if bytes,err = common.BuildResponse(-1,err.Error(),nil); err != nil {
		resp.Write(bytes)
	}
}

func handleJobList(resp http.ResponseWriter,req *http.Request) {
	var (
		jobList []*common.Job
		bytes []byte
		err error
	)
	if jobList,err = G_jobMgr.ListJobs(); err != nil {
		goto ERR
	}
	//正常应答
	if bytes,err = common.BuildResponse(0,"success",jobList);err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	if bytes,err = common.BuildResponse(-1,err.Error(),nil); err == nil {
		resp.Write(bytes)
	}
}
//强制杀死某个任务
//post /job/kill name=job1
func handleJobKill(resp http.ResponseWriter, req *http.Request){
	var (
		err error
		name string
		bytes []byte
	)
	//解析post表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}
	//要杀死的任务名
	name = req.PostForm.Get("name")
	fmt.Println(name)
	//杀死任务
	if err = G_jobMgr.KillJob(name);err!=nil {
		goto  ERR
	}
	//正常应答
	if bytes,err = common.BuildResponse(0,"success",nil);err == nil {
		resp.Write(bytes)
	}
	return
ERR:
	if bytes,err = common.BuildResponse(-1,err.Error(),nil); err == nil {
		resp.Write(bytes)
	}
}
//初始化服务
func InitApiServer() error{
	var (
		mux *http.ServeMux
		listener net.Listener
		httpServer *http.Server
		err error
		staticDir http.Dir
		staticHandler http.Handler
	)

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save",handleJobSave)
	mux.HandleFunc("/job/delete",handleJobDelete)
	mux.HandleFunc("/job/list",handleJobList)
	mux.HandleFunc("/job/kill",handleJobKill)

	//index.html

	//静态文件目录
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	mux.Handle("/",http.StripPrefix("/",staticHandler))// ./webroot/index.html

	//启动tcp监听
	if listener,err = net.Listen("tcp",":"+strconv.Itoa(G_config.ApiPort));err != nil {
		return err
	}
	//创建一个http服务
	httpServer = &http.Server{
		ReadHeaderTimeout: time.Duration(G_config.ApiReadTimeout)*time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout)*time.Millisecond,
		Handler: mux,
	}
	//赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,

	}
	go httpServer.Serve(listener)
	return nil
}
