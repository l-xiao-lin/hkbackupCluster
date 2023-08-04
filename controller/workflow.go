package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/service"
)

func WorkFlowHandler(c *gin.Context) {
	//1、获取请求参数
	p := new(model.WorkFlow)

	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("WorkFlowHandler param invalid,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}

	//2、工作流处理
	err := service.WorkFlow(p)
	zap.L().Error("service.WorkFlow failed", zap.Error(err))
	if err != nil {
		logger.SugarLog.Errorf("service.WorkFlow failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg": err.Error(),
		})
		return
	}

	//3、返回状态码
	c.JSON(200, gin.H{
		"msg":  "ASK环境部署完毕",
		"code": 1000,
	})
}
