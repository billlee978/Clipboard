package main

import (
	"ClipBoard/static"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/url"
	"os"
)

type message struct {
	Clipboard string `json:"clipboard"`
}

var messageMap = make(map[string]string)

func main() {
	r := gin.Default()

	r.NoRoute(gin.WrapH(http.FileServer(http.FS(web.Static))))

	r.POST("/upload/:token", func(c *gin.Context) {
		token := c.Param("token")
		// 单文件
		file, _ := c.FormFile("file")
		dst := "./upload/" + token
		if createTokenFolder(dst) {
			dst = dst + "/" + file.Filename
			// 上传文件至指定的完整文件路径
			err := c.SaveUploadedFile(file, dst)
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("500"))
				log.Fatalln(err)
			}
		} else {
			log.Fatalln("Create path error")
		}

		c.String(http.StatusOK, fmt.Sprintf("200"))
	})

	r.GET("download/file/:token/:fileName", func(c *gin.Context) {
		token := c.Param("token")
		fileName := c.Param("fileName")
		escapedFileName := url.PathEscape(fileName)
		// 设置响应头，告诉浏览器该文件应该被下载
		c.Header("Content-Description", "File Transfer")
		c.Header("Content-Transfer-Encoding", "binary")
		// 防止乱码
		c.Header("Content-Disposition", "attachment; filename*=utf-8''"+escapedFileName)
		c.Header("Content-Type", "application/octet-stream")
		c.File("./upload/" + token + "/" + fileName)
	})

	r.GET("delete/file/:token/:fileName", func(c *gin.Context) {
		token := c.Param("token")
		fileName := c.Param("fileName")
		file, err := os.ReadFile("./upload/" + token + "/" + fileName)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("500"))
			log.Fatalln(err)
		}
		if file != nil {
			err = os.Remove("./upload/" + token + "/" + fileName)
			if err != nil {
				c.String(http.StatusInternalServerError, fmt.Sprintf("500"))
				log.Fatalln(err)
			} else {
				c.String(http.StatusOK, fmt.Sprintf("200"))
			}
		}
	})

	r.GET("get/file/:token", func(c *gin.Context) {
		token := c.Param("token")
		dst := "./upload/" + token
		files, _ := os.ReadDir(dst)
		var filestrs []string
		for _, file := range files {
			if !file.IsDir() {
				filestrs = append(filestrs, file.Name())
			}
		}
		c.JSON(200, gin.H{
			"files": filestrs,
		})
	})

	r.POST("/add/:token", func(c *gin.Context) {
		tokenStr := c.Param("token")
		var message message
		if err := c.BindJSON(&message); err != nil {
			log.Fatalln("Read in went wrong..")
		} else {
			messageMap[tokenStr] = message.Clipboard
		}

		URL := "https://pews.top/get/" + tokenStr

		c.JSON(200, gin.H{
			"url": URL,
		})
	})

	r.GET("get/:token", func(c *gin.Context) {
		tokenStr := c.Param("token")
		c.String(200, messageMap[tokenStr])
	})

	r.Run(":8080") // 监听并在 0.0.0.0:8080 上启动服务
}

func createTokenFolder(folderPath string) bool {
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		err = os.MkdirAll(folderPath, 0755)
		if err != nil {
			return false
		} else {
			return true
		}
	}
	return true
}
