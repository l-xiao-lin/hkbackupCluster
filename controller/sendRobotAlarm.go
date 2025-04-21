package controller

import (
	"hkbackupCluster/model"
	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func SendRobotAlarmHandler(c *gin.Context) {
	p := new(model.ParamWeChatBot)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(200, gin.H{
			"msg":   "param invalid",
			"error": err.Error(),
			"code":  1001,
		})
		return
	}

	if err := service.SendRobotAlarm(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1001,
	})

}
