package callThirdPartyAPI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hkbackupCluster/logger"
	"io"
	"net/http"
	"time"
)

type paramSendImMessage struct {
	ActiveStatus    int32    `json:"activeStatus"`
	MerchantIds     []string `json:"merchantIds"`
	Message         string   `json:"message"`
	MessageSendTime string   `json:"messageSendTime"`
	OperatorName    string   `json:"operatorName"`
	TypeList        []string `json:"typeList"`
}

var message = "【系统维护通知】\n\n我们计划于10分钟后对服务器进行必要的停机升级维护，以提升服务性能和安全性。预计此次停机维护将持续5到10分钟。请您提前做好相应准备，并确保在此期间不进行关键操作。\n\n对于由此可能带来的任何不便，我们深表歉意，并感谢您的理解与支持。\n\n祝工作顺利！\n\n"

func NotifyMerchant(merchants []string) (err error) {

	apiUrl := "https://im.tongtool.com/im-service/imMessagePush/api/insert"
	data := paramSendImMessage{
		ActiveStatus:    2,
		MerchantIds:     merchants,
		Message:         message,
		MessageSendTime: time.Now().Format("2006-01-02 15:04:05"),
		OperatorName:    "系统升级",
		TypeList:        []string{"5"},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
		return
	}
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("appid", "96e79218965eb72c92a549dd5a330112")

	client := http.Client{Timeout: time.Second * 10}
	respAPI, err := client.Do(req)
	if err != nil {
		return
	}
	defer respAPI.Body.Close()

	body, err := io.ReadAll(respAPI.Body)
	if err != nil {
		return
	}
	var respData RespData

	if err := json.Unmarshal(body, &respData); err != nil {
		return err
	}
	if respData.Code == 0 && respData.Success {
		logger.SugarLog.Infof("NotifyMerchant merchants:%v success", merchants)
		return nil
	}
	logger.SugarLog.Errorf("NotifyMerchant merchants:%v failed", merchants)
	return fmt.Errorf("NotifyMerchant merchants:%v failed", merchants)
}
