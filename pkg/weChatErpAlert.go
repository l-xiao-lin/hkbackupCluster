package pkg

import "hkbackupCluster/service"

type WeChatMessage struct {
	Message string `json:"message"`
	CorpID  string `json:"corp_id"`
	Secret  string `json:"secret"`
	ToParty string `json:"toParty"`
	AgentID int    `json:"agent_id"`
}

func WeChatAlert(message string) (err error) {

	p := WeChatMessage{
		Message: message,
		CorpID:  "wxe7c550bbbe301cd3",
		Secret:  "UrUJW6Fmgdbg3vFVmssOZ6UhIThmetQeqhfmTjMVSGs",
		ToParty: "8",
		AgentID: 1000005,
	}
	//jsonData, err := json.Marshal(data)
	//if err != nil {
	//	logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
	//	return
	//}
	//resp, err := http.Post("http://127.0.0.1:8000/api/v1/sendWeChat", "application/json", bytes.NewBuffer(jsonData))
	//if err != nil {
	//	logger.SugarLog.Errorf("http.Post failed,err:%v", err)
	//	return
	//}
	//defer resp.Body.Close()

	err = service.SendWeChatAlert(p.Message, p.CorpID, p.Secret, p.ToParty, p.AgentID)
	return

}
