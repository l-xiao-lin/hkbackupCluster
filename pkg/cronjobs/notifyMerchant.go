package cronjobs

import (
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func DemoNotifyMerchantHandler(c *gin.Context) {

	err := notifyMerchant()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "package success",
		"code": 1000,
	})
}

func notifyMerchant() (err error) {
	//1、查询满足条件的记录
	records, err := mysql.GetMerchantsForIMSent()

	if len(records) == 0 {
		logger.SugarLog.Errorf("no IM sent records")
		return nil
	}

	//2、生成主机名并去重
	var hostNames []string
	var setHostMap = make(map[string]struct{})
	for _, record := range records {
		if _, exists := setHostMap[record.Host]; !exists {
			setHostMap[record.Host] = struct{}{}
			hostNames = append(hostNames, record.Host)
		}
	}

	host := strings.Join(hostNames, ":")

	//3、发送IM接口
	fmt.Printf("host:%s\n", host)

	return
}
