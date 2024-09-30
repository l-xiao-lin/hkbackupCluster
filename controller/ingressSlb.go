package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func GetIngressSlbHandler(c *gin.Context) {

	publicIP, err := service.GetIngressPublicIP()
	if err != nil {
		logger.SugarLog.Errorf("service.AddDomainRecordList failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":      "success",
		"code":     1000,
		"publicIP": publicIP,
	})
}
