package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hkbackupCluster/logger"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	tokenMutex          sync.Mutex
	tokenExpiryDuration = 20 * time.Minute
	TokenMap            = make(map[string]bool)
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

func GetAbnormalEnvName(content []string) (string, error) {
	var envNames []string
	for _, line := range content {
		parts := strings.Split(line, "]")
		if len(parts) > 1 {
			envName := strings.TrimSpace(parts[0][1:])
			envNames = append(envNames, envName)
		}
	}
	return strings.Join(envNames, ":"), nil
}

func generateToken() (string, error) {
	b := make([]byte, 10)
	_, err := rand.Reader.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil

}

func ErpErrorCount(content, host string) (err error) {
	var message string
	//处理错误数异常的环境
	resultMessage := checkAbnormalValues(content)
	if len(resultMessage) > 0 {
		envNameUrl, err := GetAbnormalEnvName(resultMessage)
		if err != nil {
			return err
		}

		//生成token
		token, err := generateToken()
		if err != nil {
			return err
		}

		//存储token的有效期
		tokenMutex.Lock()
		TokenMap[token] = true
		go func(token string) {
			time.Sleep(tokenExpiryDuration)
			tokenMutex.Lock()
			delete(TokenMap, token)
			tokenMutex.Unlock()
		}(token)
		tokenMutex.Unlock()

		restartUrl := fmt.Sprintf("http://autocheck.tongtool.com:8000/api/v1/erpRestart?envName=%s&token=%s", envNameUrl, token)

		message = fmt.Sprintf("本次发版主机为:%s,异常环境如下: \n%s是否需要重启机器 %s\n", host, strings.Join(resultMessage, ""), restartUrl)

	} else {
		message = fmt.Sprintf("本次发版主机为:%s,所有环境error数检测均正常", host)
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
