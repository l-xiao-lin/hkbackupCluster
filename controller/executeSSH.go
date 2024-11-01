package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RestarAndChecktHandler(c *gin.Context) {
	p := new(ParamRestartHost)

	if err := c.ShouldBind(p); err != nil {
		logger.SugarLog.Errorf("invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "invalid param",
			"code": 1001,
		})
		return
	}

	taskID, err := service.StartErpRestart(p.EnvName)
	if err != nil {
		logger.SugarLog.Errorf("service.StartErpRestart,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "service.StartErpRestart failed",
			"code": 1002,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
		"task": taskID,
	})

}

func CheckTaskStatusHandler(c *gin.Context) {
	taskIDStr := c.Param("task_id")
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  "invalid task id",
			"code": 1001,
		})
		return
	}

	data, err := service.CheckTaskStatus(taskID)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":    "task status",
		"code":   1000,
		"status": data.Complete,
		"error":  data.Error,
	})
}
