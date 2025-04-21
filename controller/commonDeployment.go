package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func CommonDeploymentHandler(c *gin.Context) {
	p := new(model.ParamCommonDeploy)
	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("CommonDeploymentHandler invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}

	logger.SugarLog.Infof("CommonDeploymentHandler param:%+v", *p)
	if err := service.CommonDeployment(p); err != nil {
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
