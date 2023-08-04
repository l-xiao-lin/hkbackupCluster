package controller

import (
	"github.com/gin-gonic/gin"
	"hkbackupCluster/service"
)

func ExecCommandHandler(c *gin.Context) {
	paramCmd := c.Query("cmd")
	paramHost := c.Query("host")

	data, err := service.ExecCommand(paramCmd, paramHost)
	if err != nil {
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "Command executed successfully",
		"code": 1000,
		"data": data,
	})

}
