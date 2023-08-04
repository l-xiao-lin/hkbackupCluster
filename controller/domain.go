package controller

import (
	"github.com/gin-gonic/gin"
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
)

func AddDomainRecordHandler(c *gin.Context) {
	p := new(ParamsDomainRecord)
	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("AddDomainRecordHandler invalid parama,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 10001,
		})
		return
	}

	err := service.AddDomainRecordList(p.AccessKeyId, p.AccessKeySecret, p.Value, p.Records)
	if err != nil {
		logger.SugarLog.Errorf("service.AddDomainRecordList failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "添加A记录成功",
		"code": 1000,
	})
}
