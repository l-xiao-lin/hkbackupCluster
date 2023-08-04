package sshclient

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"hkbackupCluster/logger"
	"io/ioutil"
)

func SshConnect(remoteHost string) (client *ssh.Client, err error) {
	key, err := ioutil.ReadFile("./conf/id_rsa")
	if err != nil {
		logger.SugarLog.Errorf("ioutil.ReadFile failed,err:%v\n", err)
		return
	}

	//解析密钥
	singer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		logger.SugarLog.Errorf("failed to ParsePrivateKey,err: %v\n", err)
		return
	}

	//ssh连接配置
	config := &ssh.ClientConfig{
		User: "tomcat",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(singer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	//连接远程主机
	client, err = ssh.Dial("tcp", fmt.Sprintf("%s:22", remoteHost), config)
	if err != nil {
		logger.SugarLog.Errorf("Failed to dial,err:%v\n", err)
		return nil, err
	}
	return client, nil
}
