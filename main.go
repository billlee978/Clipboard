package main

import (
	"ClipBoard/static"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
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
		fileDst := "./upload/" + token + "/" + fileName
		file, err := os.Open(fileDst)
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		defer file.Close()
		stat, err := file.Stat()
		if err != nil {
			c.AbortWithError(404, err)
			return
		}
		// 设置响应头，告诉浏览器该文件应该被下载
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Header().Set("Content-Length", strconv.FormatInt(stat.Size(), 10))
		c.Writer.Flush()
		var offset int64 = 0
		var bufsize int64 = 1024 * 1024 // 1MB
		buf := make([]byte, bufsize)
		for {
			n, err := file.ReadAt(buf, offset)
			if err != nil && err != io.EOF {
				log.Println("read file error", err)
				break
			}
			if n == 0 {
				break
			}
			_, err = c.Writer.Write(buf[:n])
			if err != nil {
				log.Println("write file error", err)
				break
			}
			offset += int64(n)
		}
		c.Writer.Flush()
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

		URL := "http://" + localAddr + ":8080/get/" + tokenStr

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
