package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hkbackupCluster/pkg"
	"hkbackupCluster/service"
)

func CheckListingHandler(c *gin.Context) {
	p := new(ParamListing)
	if err := c.ShouldBind(p); err != nil {
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}
	var message string
	err := service.CheckListing(p.Website, p.Username, p.Password)
	if err != nil {
		message = fmt.Sprintf("Failed 环境:%s检测失败,err:%v", p.Website, err)
		pkg.ListingWeChatAlert(message)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	message = fmt.Sprintf("Success 环境:%s检测通过", p.Website)
	pkg.ListingWeChatAlert(message)
	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}
