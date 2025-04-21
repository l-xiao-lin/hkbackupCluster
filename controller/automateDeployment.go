package controller

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AutomateDeploymentHandler(c *gin.Context) {
	//1、初始化默认参数
	p := new(model.ParamsIncrementalPack)
	//2、参数校验

	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("IncrementalPackHandler invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}

	logger.SugarLog.Infof("AutomateDeploymentHandler params:%+v", *p)

	//2、业务处理 ,
	insertID, err := service.AutomateDeployment(p)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  fmt.Sprintf("package failed,err:%v", err),
			"code": 1002,
		})
		return
	}

	//3、状态返回
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
		"data": map[string]int64{
			"insertID": insertID,
		},
	})
}

func UpdateStatusHandler(c *gin.Context) {
	p := new(model.ParamsUpdateStatus)
	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("UpdateStatusHandler invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}

	if err := service.UpdateStatus(p); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  fmt.Errorf("service.UpdateStatus,err:%v", err),
			"code": 1002,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}
