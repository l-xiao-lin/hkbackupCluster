package cronjobs

import (
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"net/http"
	"time"

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
		"msg":  "notify merchant success",
		"code": 1000,
	})
}

func notifyMerchant() (err error) {
	//1、查询满足条件的记录
	logger.SugarLog.Infof("begin notifyMerchant func at %s", time.Now().Format("2006-01-02 15:04:05"))
	records, err := mysql.GetMerchantsForIMSent()

	if len(records) == 0 {
		logger.SugarLog.Infof("no IM sent records")
		return nil
	}

	//2、生成主机切片并将主机元素去重
	uniqueHostMap := make(map[string]bool)
	for _, record := range records {
		hostList, err := generationHosts(record.Host)
		if err != nil {
			logger.SugarLog.Errorf("generationHosts failed,taskID:%s host:%s", record.TaskID, record.Host)
			continue
		}
		var uniqueHostSlice []string
		for _, host := range hostList {
			if _, exists := uniqueHostMap[host]; !exists {
				uniqueHostMap[host] = true
				uniqueHostSlice = append(uniqueHostSlice, host)
			}
		}
		logger.SugarLog.Infof("uniqueHostSlice :%s", uniqueHostSlice)
		var imSentStatus int
		if len(uniqueHostSlice) > 0 {
			//一、将主机转换成商户号
			merchants, err := callThirdPartyAPI.GetMerchantsByEnvHost(uniqueHostSlice)
			if err != nil {
				logger.SugarLog.Errorf("GetMerchantsByEnvHost failed,taskID:%s,err:%v", record.TaskID, err)
				imSentStatus = 2 //发送IM通知失败
				if err := mysql.UpdateIMSentStatus(imSentStatus, record.TaskID); err != nil {
					logger.SugarLog.Errorf("UpdateIMSentStatus failed,taskID:%s,err:%v", record.TaskID, err)
				}
				continue
			}

			//二、调用IM通知接口
			if err := callThirdPartyAPI.NotifyMerchant(merchants); err != nil {
				imSentStatus = 2
				if err := mysql.UpdateIMSentStatus(imSentStatus, record.TaskID); err != nil {
					logger.SugarLog.Errorf("UpdateIMSentStatus failed,taskID:%s,err:%v", record.TaskID, err)
				}
				continue
			}
			imSentStatus = 1 //发送IM成功
			if err := mysql.UpdateIMSentStatus(imSentStatus, record.TaskID); err != nil {
				logger.SugarLog.Errorf("UpdateIMSentStatus failed,taskID:%s,err:%v", record.TaskID, err)
			}
			logger.SugarLog.Infof("UpdateIMSentStatus success,taskID:%s", record.TaskID)
		}
	}

	return
}
