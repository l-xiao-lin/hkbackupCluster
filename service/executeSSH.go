package service

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/sshNewClient"
	"sync"
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
	wg            sync.WaitGroup
	taskIDCounter int
	tasks         = make(map[int]*TaskStatus)
)

type TaskStatus struct {
	ID       int   `json:"id"`
	Complete bool  `json:"complete"`
	Error    error `json:"error"`
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
	time.Sleep(retryInterval * checkRetryCount)

	errCheckCommand := fmt.Sprintf("~/ansible/scripts/check_erp_log.sh %s 2", host)

	if err = executeCommand(errCheckCommand); err != nil {
		return err
	}
	return nil
}

func StartErpRestart(host string) (int, error) {
	//生成唯一的任务ID
	m.Lock()
	taskIDCounter++
	taskID := taskIDCounter
	tasks[taskID] = &TaskStatus{
		ID: taskID,
	}
	m.Unlock()

	//开始执行后台任务
	wg.Add(1)
	go func(host string, taskID int) {
		defer wg.Done()
		err := RestartAndCheck(host)
		m.Lock()
		tasks[taskID].Complete = true
		tasks[taskID].Error = err
		m.Unlock()
	}(host, taskID)
	return taskID, nil
}

func CheckTaskStatus(taskID int) (*TaskStatus, error) {
	m.Lock()
	defer m.Unlock()
	status, ok := tasks[taskID]
	if !ok {
		logger.SugarLog.Errorf("task with ID %d not found", taskID)
		return nil, fmt.Errorf("task with ID %d not found", taskID)
	}
	return status, nil
}
