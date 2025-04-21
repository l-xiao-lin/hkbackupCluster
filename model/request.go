package model

type ParamSSHConfig struct {
	Host    string `json:"host"`
	User    string `json:"user"`
	Port    int    `json:"port"`
	KeyPath string `json:"keyPath"`
	Command string `json:"command"`
}

type ParamWeChatBot struct {
	Key                 string   `json:"Key"`
	MsgType             string   `json:"msgtype"`
	Content             string   `json:"content"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}
