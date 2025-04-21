package controller

import (
	"hkbackupCluster/model"
	"hkbackupCluster/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func TestEnvPackageHandler(c *gin.Context) {
	p := new(model.ParamTestPack)
	if err := c.ShouldBindJSON(p); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}

	go func(param *model.ParamTestPack) {
		service.TestEnvPackage(param)
	}(p)

	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}
