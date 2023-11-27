package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func SendWeChatHandler(c *gin.Context) {
	p := new(ParamWeChat)
	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("SendWeChatHandler ShouldBindJSON failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定错误",
			"code": 1001,
		})
		return
	}

	if err := service.SendWeChatAlert(p.Message, p.CorpID, p.Secret, p.ToParty, p.AgentID); err != nil {
		logger.SugarLog.Errorf("service.SendWeChatAlert,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	logger.SugarLog.Info("service.SendWeChatAlert success")
	c.JSON(200, gin.H{
		"msg":  "send WeChat success",
		"code": 1000,
	})
}
