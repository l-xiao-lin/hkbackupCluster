package sshNewClient

import (
	"bytes"
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"os"

	"golang.org/x/crypto/ssh"
)

func ExecuteSSHCommand(conf *model.ParamSSHConfig) (string, error) {
	privateBytes, err := os.ReadFile(conf.KeyPath)
	if err != nil {
		logger.SugarLog.Errorf("os.ReadFile failed,err:%v", err)
		return "", err
	}
	signer, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		logger.SugarLog.Errorf("ssh.ParsePrivateKey failed,err:%v", err)
		return "", err
	}

	connConfig := &ssh.ClientConfig{
		User: conf.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	configStr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
	conn, err := ssh.Dial("tcp", configStr, connConfig)
	if err != nil {
		logger.SugarLog.Errorf("ssh.Dial failed,err:%v", err)
		return "", err
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		logger.SugarLog.Errorf("conn.NewSession failed,err:%v", err)
		return "", err
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	err = session.Run(conf.Command)
	if err != nil {
		logger.SugarLog.Errorf("session.Run failed,err:%v", err)
		return "", err
	}
	output := stdoutBuf.String()
	return output, nil

}
