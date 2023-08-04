package service

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/sshclient"
)

func ExecCommand(cmd, remoteHost string) (ResponseData string, err error) {

	client, err := sshclient.SshConnect(remoteHost)
	if err != nil {
		return
	}
	defer client.Close()

	//创建一个session
	session, err := client.NewSession()
	if err != nil {
		logger.SugarLog.Errorf("Failed to create session,err:%v\n", err)
		return
	}
	defer session.Close()

	//执行远程命令
	cmdInfo, err := session.CombinedOutput(cmd)
	if err != nil {
		return
	}
	return string(cmdInfo), nil

}
