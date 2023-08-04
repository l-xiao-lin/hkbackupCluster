package controller

import (
	"github.com/gin-gonic/gin"
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
)

func UploadFileHandler(c *gin.Context) {
	p := new(ParamUpload)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  "参数绑定错误",
			"code": 1002,
		})
		return
	}

	if err := service.UploadFile(p.LocalFilePath, p.RemoteDir, p.RemoteHost); err != nil {
		logger.SugarLog.Errorf("service.UploadFile failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 10002,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "file upload success",
		"code": 1000,
	})

}
