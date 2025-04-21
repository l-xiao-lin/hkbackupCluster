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

type RespData struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Success bool   `json:"success"`
	Datas   int    `json:"datas"`
}

type param struct {
	ReleaseTaskID string `json:"releaseTaskId"`
}

func TriggerSQLExecution(taskID string) (err error) {
	apiUrl := "http://10.0.0.180:8138/releaseTask/api/noticeExecuteSql"
	data := param{ReleaseTaskID: taskID}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
		return
	}
	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.SugarLog.Errorf("http.NewRequest failed,err:%v", err)
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
	if respData.Code != 0 {
		logger.SugarLog.Errorf("call callThirdPartyAPI failed,err:%v", respData.Message)
		return fmt.Errorf("TriggerSQLExecution failed,taskID:%s,err:%v", taskID, respData.Message)
	}
	logger.SugarLog.Infof("TriggerSQLExecution taskID:%s success.", taskID)
	return nil

}
