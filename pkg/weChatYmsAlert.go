package pkg

import "hkbackupCluster/service"

func WeChatYmsAlert(message string) (err error) {

	p := WeChatMessage{
		Message: message,
		CorpID:  "wxe7c550bbbe301cd3",
		Secret:  "kI8WxQaITnZfwo0w3NcxvM-IpIcnYC7YWMM0SCUtVoA",
		ToParty: "7",
		AgentID: 1000002,
	}
	err = service.SendWeChatAlert(p.Message, p.CorpID, p.Secret, p.ToParty, p.AgentID)
	return
}
