package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg"
	"hkbackupCluster/service"
	"hkbackupCluster/settings"
	"strings"
)

func CheckSystemHandler(c *gin.Context) {

	p := new(ParamCheck)

	err := c.ShouldBind(p)
	envNames := strings.Split(p.EnvName, ",")
	if err != nil || len(envNames) == 0 {
		logger.SugarLog.Errorf("CheckSystemHandler ShouldBindJSON failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定错误",
			"code": 1001,
		})
		return
	}

	var message string
	//将环境名转换成商户号
	merchantsID, invalidEnv := settings.EnvFindMerchant(envNames)
	if len(invalidEnv) > 0 {
		//发送微信告警
		message = fmt.Sprintf("无效的环境名,data:%v", invalidEnv)
		pkg.WeChatAlert(message)

		c.JSON(200, gin.H{
			"msg":  "无效的环境名",
			"code": 1003,
			"data": invalidEnv,
		})
		return
	}

	data := service.CheckSystem(merchantsID)
	if len(data) > 0 {
		//获取map中需要失败的merchantsID
		var errMerchants []string
		for errMerchant := range data {
			errMerchants = append(errMerchants, errMerchant)
		}

		errEnvNames := settings.MerchantFindEnv(errMerchants)

		logger.SugarLog.Errorf("service.CheckSystem errors:%v", data)

		message = fmt.Sprintf("Failed %v环境检测失败", errEnvNames)
		pkg.WeChatAlert(message)

		c.JSON(200, gin.H{
			"msg":  "内部错误",
			"code": 1002,
			"data": fmt.Sprintf("Failed %v环境检测失败", errEnvNames),
		})
		return
	}

	logger.SugarLog.Info("service.CheckSystem success")

	message = fmt.Sprintf("Success 环境检测通过,共检测%d个环境", len(envNames))
	pkg.WeChatAlert(message)

	c.JSON(200, gin.H{
		"msg":  fmt.Sprintf("Success 环境检测通过"),
		"code": 1000,
	})

}
