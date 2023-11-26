package main

import (
	"ClipBoard/static"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid2 "github.com/google/uuid"
	"log"
	"net"
	"net/http"
	"strings"
)

type message struct {
	Clipboard string `json:"clipboard"`
}

var messageMap = make(map[string]string)
var localAddr string

func init() {
	var ips []string = GetIps()
	for _, ip := range ips {
		if strings.HasPrefix(ip, "172") {
			localAddr = ip
			break
		}
	}
}

func main() {
	r := gin.Default()

	r.NoRoute(gin.WrapH(http.FileServer(http.FS(web.Static))))

	r.POST("/add", func(c *gin.Context) {
		var message message
		uuid := uuid2.NewString()
		if err := c.BindJSON(&message); err != nil {
			log.Fatalln("Read in went wrong..")
		} else {
			messageMap[uuid] = message.Clipboard
		}

		URL := "http://" + localAddr + ":8080/get/" + uuid

		c.JSON(200, gin.H{
			"url": URL,
		})
	})

	r.GET("get/:uuid", func(c *gin.Context) {
		uuid := c.Param("uuid")
		c.String(200, messageMap[uuid])
	})

	r.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}

func GetIps() (ips []string) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interfaces ipAddress: %v\n", err)
		return ips
	}

	for _, address := range interfaceAddr {
		ipNet, isVailIpNet := address.(*net.IPNet)
		// 检查ip地址判断是否回环地址
		if isVailIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP.String())
			}
		}
	}
	return ips
}
