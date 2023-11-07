package controller

import (
	"github.com/gin-gonic/gin"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/service"
)

func CompareJarHandler(c *gin.Context) {
	p := new(model.ParamsCompareJar)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 10001,
		})
		return
	}
	data, err := service.CompareJar(p)
	if err != nil {
		logger.SugarLog.Errorf("service.CompareJar,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 10002,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "compare jar success...",
		"data": data,
		"code": 1000,
	})
}
