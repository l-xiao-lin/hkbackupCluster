package service

import (
	"hkbackupCluster/pkg/sshclient"
	"strings"
)

var ansibleHost = "172.16.60.1"

func GetEnvName(cmd string) (RespData []string, err error) {

	client, err := sshclient.SshConnect(ansibleHost)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmdInfo, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(cmdInfo)), "\n")
	for _, line := range lines {
		RespData = append(RespData, strings.TrimSpace(line))

	}
	RespData = append(RespData, "standalone:guanwang:guanwang-i2:sdk")
	RespData = append(RespData, "monday", "wednesday")
	RespData = append(RespData, "standalone:guanwang:guanwang-i2:sdk:!monday")
	RespData = append(RespData, "standalone:guanwang:guanwang-i2:sdk:!wednesday")

	return RespData, nil
}

func GetInventory(cmd string) (RespData []string, err error) {

	client, err := sshclient.SshConnect(ansibleHost)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmdInfo, err := session.CombinedOutput(cmd)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(cmdInfo)), "\n")
	for _, line := range lines {
		RespData = append(RespData, strings.TrimSpace(line))

	}
	return RespData, nil
}
