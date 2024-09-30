package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

func readFile(file *multipart.FileHeader) (string, error) {
	uploadedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer uploadedFile.Close()

	//读取内容
	var content []byte
	buffer := make([]byte, 1024)
	for {
		n, err := uploadedFile.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil && err != io.EOF {
			return "", nil
		}
		content = append(content, buffer[:n]...)

	}
	return string(content), nil
}

func ErpErrorCountHandler(c *gin.Context) {
	//从请求中获取名为file的文件
	file, err := c.FormFile("file")
	if err != nil {
		logger.SugarLog.Error("ErpErrorCountHandler invalid param")
		c.JSON(200, gin.H{
			"msg":  "invalid param",
			"code": 1001,
		})
		return
	}
	//读取文件内容
	content, err := readFile(file)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  "read file failed",
			"code": 1002,
		})
		return
	}
	//业务处理
	if err = service.ErpErrorCount(content); err != nil {
		c.JSON(200, gin.H{
			"msg":  "service.ErpErrorCount failed",
			"code": 1004,
		})
		return
	}

	//状态返回
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})
}
