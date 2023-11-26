package main

import (
	"ClipBoard/static"
	"github.com/gin-gonic/gin"
	uuid2 "github.com/google/uuid"
	"log"
	"net/http"
)

type message struct {
	Clipboard string `json:"clipboard"`
}

var messageMap = make(map[string]string)

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

		URL := "http://localhost:8080/get/" + uuid

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
