package service

import (
	"fmt"
	"hkbackupCluster/logger"
	"strconv"
	"strings"
)

type WeChatMessageErp struct {
	Message string `json:"message"`
	CorpID  string `json:"corp_id"`
	Secret  string `json:"secret"`
	ToParty string `json:"toParty"`
	AgentID int    `json:"agent_id"`
}

func checkAbnormalValues(content string) []string {
	lines := strings.Split(content, "\n")
	normalValues := []int{-1, 0, 1, 32, 63, 94, 125, 156, 187, 218}
	var abnormalInfo []string
	for _, line := range lines {
		if strings.Contains(line, "]") {
			parts := strings.Split(line, "]")
			if len(parts) > 1 {
				valuePart := strings.TrimSpace(parts[1])
				value, err := strconv.Atoi(valuePart)
				if err != nil {
					logger.SugarLog.Errorf("Failed to parse value:%v", valuePart)
					continue
				}
				isNormal := false

				for _, normalValue := range normalValues {
					if value == normalValue {
						isNormal = true
						break
					}

				}
				if !isNormal {

					abnormalInfo = append(abnormalInfo, fmt.Sprintf("%s\n", line))
				}

			}
		}
	}
	return abnormalInfo

}

func ErpErrorCount(content string) (err error) {
	var message string
	//处理错误数异常的环境
	resultMessage := checkAbnormalValues(content)
	if len(resultMessage) > 0 {
		message = fmt.Sprintf("本次发版异常环境如下: \n%s", strings.Join(resultMessage, ""))
	} else {
		message = fmt.Sprintf("本次发版所有环境error数检测均正常")
	}

	fmt.Println(message)
	//初始化企业微信配置
	p := WeChatMessageErp{
		Message: message,
		CorpID:  "wxe7c550bbbe301cd3",
		Secret:  "UrUJW6Fmgdbg3vFVmssOZ6UhIThmetQeqhfmTjMVSGs",
		ToParty: "8",
		AgentID: 1000005,
	}

	//发送微信

	return SendWeChatAlert(p.Message, p.CorpID, p.Secret, p.ToParty, p.AgentID)
}
