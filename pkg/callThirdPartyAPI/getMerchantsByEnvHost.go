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

type EnvHosts struct {
	EnvironmentCodeList []string `json:"environmentCodeList"`
}

func GetMerchantsByEnvHost(envHosts []string) (merchants []string, err error) {
	apiUrl := "https://automated.pvt.tongtool.com/tool-service/environment/api/getMerchantByEnvCodes"
	data := EnvHosts{EnvironmentCodeList: envHosts}
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
		return nil, err
	}

	if respData.Code == 0 && respData.Datas != nil {
		logger.SugarLog.Infof("GetMerchantsByEnvHost success.")
		return respData.Datas, nil
	}
	logger.SugarLog.Errorf("call GetMerchantsByEnvHost failed,err:%v", respData.Message)
	return nil, fmt.Errorf("GetMerchantsByEnvHost failed,taskID:%s", envHosts)
}
