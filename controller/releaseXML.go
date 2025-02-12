package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func ReleaseXmlHandler(c *gin.Context) {
	p := new(model.ParamReleaseXML)
	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("ReleaseXmlHandler invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}

	if err := service.ReleaseXml(p); err != nil {
		logger.SugarLog.Errorf("service.ReleaseXml failed,taskID:%s,common:%s,err:%v", p.TaskID, *p.Common, err)
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
