package controller

import (
	"fmt"
	"hkbackupCluster/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

var defaultHost = "standalone:guanwang:guanwang-i2:sdk"

func GetEnvNameHandler(c *gin.Context) {

	host := c.DefaultQuery("host", defaultHost)

	cmd := fmt.Sprintf("ansible -i $Z_asbList_erp2 %s --list| grep -v hosts", host)

	data, err := service.GetEnvName(cmd, host)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Command executed successfully",
		"code": 1000,
		"data": data,
	})
}

func GetInventoryHandler(c *gin.Context) {
	group := c.Query("group")
	if group == "" {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Group parameter is required",
			"code": 1001,
		})
		return

	}
	fmt.Printf("group:%s", group)

	cmd := fmt.Sprintf("ansible -i $Z_asbList_erp2 %s --list| grep -v hosts", group)

	data, err := service.GetInventory(cmd)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "Command executed successfully",
		"code": 1000,
		"data": data,
	})
}
