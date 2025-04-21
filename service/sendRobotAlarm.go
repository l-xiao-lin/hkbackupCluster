package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"net/http"
)

type BotText struct {
	Content             string   `json:"content"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type WeChatBot struct {
	MsgType string  `json:"msgtype"`
	Text    BotText `json:"text"`
}

func SendRobotAlarm(p *model.ParamWeChatBot) (err error) {
	param := WeChatBot{MsgType: p.MsgType, Text: BotText{
		Content:             p.Content,
		MentionedMobileList: p.MentionedMobileList,
	}}
	payload, err := json.Marshal(param)
	logger.SugarLog.Infof("param json:%s", string(payload))
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
		return
	}
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=%s", p.Key)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		logger.SugarLog.Errorf("http.Post failed,err:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logger.SugarLog.Errorf("resp.StatusCode not statusOk")
		return
	}
	return nil

}
