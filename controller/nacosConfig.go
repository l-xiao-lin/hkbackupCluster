package controller

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
	"io"
)

func GetConfigHandler(c *gin.Context) {
	p := new(ParamConf)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}

	err := service.GetConfig(p.Namespace, p.DataID, p.Group)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})
}

func PublishConfigHandler(c *gin.Context) {
	p := new(ParamConf)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}
	//获取上传的文件
	file, err := c.FormFile("filename")
	if err != nil {
		logger.SugarLog.Errorf("get config failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	src, err := file.Open()
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1003,
		})
		return
	}

	defer src.Close()
	var buf bytes.Buffer
	var content string
	_, err = io.Copy(&buf, src)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1004,
		})
		return
	}
	content = buf.String()

	err = service.PublishConfig(p.Namespace, p.DataID, p.Group, content)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})
}
