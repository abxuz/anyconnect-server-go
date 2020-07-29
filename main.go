package main

import (
	"acserv/server"
	"crypto/tls"
	"log"
	"net/http"
	"os"
)

func main() {

	// 1. 读取配置
	// 2. 初始化组件
	// 3. 运行服务
	cert, err := tls.LoadX509KeyPair("server.pem", "server.pem")
	if err != nil {
		log.Fatal(err)
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	tlsConfig.BuildNameToCertificate()

	server := &server.Server{
		HTMLFileSystem: http.Dir("html"),
		Loger:          log.New(os.Stdout, "", log.LstdFlags),
	}
	server.ListenAndServeTLS("tcp", ":https", tlsConfig)
}
