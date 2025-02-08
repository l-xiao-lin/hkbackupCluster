package controller

import (
	"fmt"
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

	// 启动一个新的 goroutine 处理耗时任务
	go func(param *model.ParamReleaseXML) {
		if err := service.ReleaseXml(p); err != nil {
			c.JSON(200, gin.H{
				"msg":  fmt.Sprintf("service.ReleaseXml failed,err:%v", err),
				"code": 1002,
			})
			return
		}
	}(p)

	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}

func GetXmlStatusHandler(c *gin.Context) {
	taskID := c.Query("task_id")
	status, exists := service.GetXmlTaskStatus(taskID)
	if !exists {
		c.JSON(200, gin.H{
			"msg":  "Task not found",
			"code": 1005,
		})
		return
	}
	c.JSON(200, gin.H{
		"code":   1000,
		"status": status.Status,
		"error":  status.Error,
	})
}
