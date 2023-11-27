package service

import (
	"bytes"
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"net/http"

	"github.com/goccy/go-json"
)

var ErrTokenInvalid = errors.New("invalid token")

type WeChatMessage struct {
	ToParty string `json:"toparty"`
	MsgType string `json:"msgtype"`
	Text    *Text  `json:"text"`
	AgentID int    `json:"agentid"`
}

type Text struct {
	Content string `json:"content"`
}

func SendWeChatAlert(message, corpID, secret, toParty string, agentID int) error {
	token, err := getAccessToken(corpID, secret)
	if err != nil {
		logger.SugarLog.Errorf("getAccessToken failed,err:%v", err)
		return err
	}

	data := WeChatMessage{
		ToParty: toParty,
		MsgType: "text",
		Text:    &Text{Content: message},
		AgentID: agentID,
	}
	payload, err := json.Marshal(data)
	fmt.Println(string(payload))
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal,err:%v", err)
		return err
	}

	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s", token)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		logger.SugarLog.Errorf("http.Post failed,err:%v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil

}

func getAccessToken(corpID, secret string) (token string, err error) {
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s", corpID, secret)
	resp, err := http.Get(apiURL)
	if err != nil {
		logger.SugarLog.Errorf("http GET failed,err:%v", err)
		return "", err
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if result["access_token"] == nil {
		logger.SugarLog.Errorf("token is invalid")
		return "", ErrTokenInvalid
	}
	fmt.Println(result)
	token = result["access_token"].(string)

	return token, nil
}
