package main

import (
	"net/http"
	"strings"
	"crypto/md5"
	"time"
	"fmt"
	"os"
	"io"
	"path/filepath"
	"encoding/json"
)

// 第1步 编写handler
// 初体验. 输出hello world
func sayHello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

// 首页html文件
func indexHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./main.html")
}

func main()  {
	// 第2步 将handler注册进servermux  就是将不同url的请求交给对应的handler处理
	// 初体验. 输出hello world
	http.HandleFunc("/sayHello", sayHello)

	// 首页html文件
	http.HandleFunc("/", indexHandler)

	// 实现video静态文件接口
	fileHandler := http.FileServer(http.Dir("./video")) //静态文件服务器的目录为video
	// 第一个video是指访问的url为http://xxx/video/mp4, 第二个video为将url重写为 mp4
	// 因此只剩下mp4这个文件名, 正好为静态文件服务器目录下的mp4文件; 所以这里http.Dir("./video")中的video和下面两个video没关系
	http.Handle("/video/", http.StripPrefix("/video/", fileHandler))


	// 注册上传文件的handler
	http.HandleFunc("/api/upload", uploadHandler)

	// 注册获取视频list的handler
	http.HandleFunc("/api/list", getFileListHandler)


	// 第3步: 启动web服务
	http.ListenAndServe(":8090", nil)
}

// 1.上传视频文件的业务逻辑
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1.限制客户端上传视频文件的大小， 最大为 10*1024*1024B = 10*1024KB = 10MB
	r.Body = http.MaxBytesReader(w, r.Body, 10 * 1024 * 1024)
	// 服务器最多只接收10MB数据, 如果文件超过10MB会被截断, 这里请求就会判断为非合法的结构
	err := r.ParseMultipartForm(10 * 1024 * 1024)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 2.获取上传的文件
	file, fileHeader, err := r.FormFile("uploadFile")

	// 3.检查文件类型, 文件名称的后缀必须为.mp4
	ret := strings.HasSuffix(fileHeader.Filename, ".mp4")
	if ret == false {
		http.Error(w, "not mp4", http.StatusInternalServerError)
		return
	}

	// 4.获取随机名称, 防止用户上传了同名文件造成文件覆盖
	md5Byte := md5.Sum([]byte(fileHeader.Filename + time.Now().String()))
	md5Str := fmt.Sprintf("%x", md5Byte)
	newFileName := md5Str + ".mp4"

	// 5.写入文件
	dst, err := os.Create("./video/"+newFileName)
	defer  dst.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

// 获取视频文件列表
func getFileListHandler(w http.ResponseWriter, r *http.Request)  {
	// 获取video文件夹下的所有文件
	files, _ := filepath.Glob("video/*")
	var ret []string
	for _, file := range files {
		ret = append(ret, "http://" + r.Host + "/video/" + filepath.Base(file))
	}
	retJson, _ := json.Marshal(ret)
	w.Write(retJson)
	return
}