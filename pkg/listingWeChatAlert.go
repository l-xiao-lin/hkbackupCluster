package pkg

import (
	"bytes"
	"encoding/json"
	"hkbackupCluster/logger"
	"net/http"
)

func ListingWeChatAlert(message string) (err error) {

	data := WeChatMessage{
		Message: message,
		CorpID:  "wxe7c550bbbe301cd3",
		Secret:  "YAJ9bAbT6unWXZyt1up3E5FVmNOA3d_2QaD9xuFFWtE",
		ToParty: "9",
		AgentID: 1000006,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
		return
	}
	resp, err := http.Post("http://127.0.0.1:8000/api/v1/sendWeChat", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.SugarLog.Errorf("http.Post failed,err:%v", err)
		return
	}
	defer resp.Body.Close()
	return
}
