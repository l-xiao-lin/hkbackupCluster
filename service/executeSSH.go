package service

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/sshNewClient"
	"time"
)

const (
	defaultHost     = "172.16.60.1"
	defaultUser     = "tomcat"
	defaultPort     = 22
	defaultKeyPath  = "./conf/tomcat_bastion3"
	retryInterval   = 60 * time.Second
	checkRetryCount = 4
)

var (
	taskIDCounter int
	tasks         = make(map[int]*taskStatus)
)

type taskStatus struct {
	ID       int
	Complete bool
	Error    error
}

func executeCommand(command string) error {
	conf := &model.ParamSSHConfig{
		Host:    defaultHost,
		User:    defaultUser,
		Port:    defaultPort,
		KeyPath: defaultKeyPath,
		Command: command,
	}
	response, err := sshNewClient.ExecuteSSHCommand(conf)
	if err != nil {
		logger.SugarLog.Errorf("sshNewClient.ExecuteSSHCommand failed,err:%v", err)
		return err
	}
	logger.SugarLog.Info("Response from command execution :%s", response)
	return nil
}

func RestartAndCheck(host string) (err error) {
	//1、初始化ansible参数

	command := fmt.Sprintf("~/.pyenv/versions/ansible-2.7/bin/ansible-playbook  /home/tomcat/ansible/deploy/stop_start_jboss_sdk.yml -i /home/tomcat/ansible/deploy/Inventory/wanip_hosts -e \"host=%s restart_jboss=true\" -t rs_jboss", host)

	//2、调用重启命令

	if err = executeCommand(command); err != nil {
		return err
	}

	//3、等待4分钟后，再执行错误检测命令
	for i := 0; i < checkRetryCount; i++ {
		time.Sleep(retryInterval)
		fmt.Printf("time sleep %d minute\n", i)
	}

	errCheckCommand := fmt.Sprintf("~/ansible/scripts/check_erp_log.sh %s 2", host)

	if err = executeCommand(errCheckCommand); err != nil {
		return err
	}
	return nil
}

func StartErpRestart(host string) (int, error) {
	//1、生成taskID
	m.Lock()
	taskIDCounter++
	taskID := taskIDCounter
	tasks[taskID] = &taskStatus{
		ID: taskID,
	}
	m.Unlock()

	//2、执行后台重启及检测操作
	go func(host string, taskID int) {
		err := RestartAndCheck(host)
		m.Lock()
		tasks[taskID].Complete = true
		tasks[taskID].Error = err
		m.Unlock()
	}(host, taskID)

	//3、返回给用户taskID信息
	return taskID, nil
}

func CheckTaskStatus(taskID int) (*taskStatus, error) {
	m.Lock()
	defer m.Unlock()
	status, ok := tasks[taskID]
	if !ok {
		logger.SugarLog.Errorf("task with ID %d not found", taskID)
		return nil, fmt.Errorf("task with ID %d not found", taskID)
	}
	return status, nil

}
