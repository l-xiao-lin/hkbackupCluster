package service

import (
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"hkbackupCluster/pkg/sshclient"
	"strings"
	"time"
)

var (
	jenkinsMaster   = "10.0.0.130"
	buildXMLJobName = "iBizXML-通过源文件自动获取目标路径"
)

type paramBuildXML struct {
	Host     string `json:"host"`
	SrcFile  string `json:"src_file"`
	DestPath string `json:"dest_path"`
}

type XmlTaskStatus struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

var xmlTaskStatusMap = make(map[string]XmlTaskStatus)

func UpdateXmlTaskStatus(taskID string, status string, err error) {
	m.Lock()
	defer m.Unlock()
	if err != nil {
		xmlTaskStatusMap[taskID] = XmlTaskStatus{Status: status, Error: err.Error()}
	} else {
		xmlTaskStatusMap[taskID] = XmlTaskStatus{Status: status, Error: ""}
	}

	//清理过过期的任务状态
	go func() {
		time.Sleep(10 * time.Minute)
		m.Lock()
		delete(xmlTaskStatusMap, taskID)
		m.Unlock()
	}()
	logger.SugarLog.Infof("xmlTaskStatusMap:%v", xmlTaskStatusMap)
}

// GetXmlTaskStatus 获取任务状态
func GetXmlTaskStatus(taskID string) (XmlTaskStatus, bool) {
	fmt.Printf("GetXmlTaskStatus func xmlTaskStatusMap:%v ", xmlTaskStatusMap)
	m.Lock()
	defer m.Unlock()
	status, exists := xmlTaskStatusMap[taskID]
	return status, exists

}

func getXMLPathByBugNumber(bugNumber string) (RespData []string, err error) {
	client, err := sshclient.SshConnect(jenkinsMaster)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	cmd := fmt.Sprintf("find /home/tomcat/ansible/src/add/  -type f  ! -name \"*.class\" | grep iBizXML| grep %s", bugNumber)
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

func ReleaseXml(p *model.ParamReleaseXML) (err error) {
	var failedTasks []string
	//1、调用jenkins(拷贝-当天增量-到堡垒机)
	resp, err := pack.JenkinsBuild(p.JobName, nil)
	if err != nil || resp.BuildResult != "SUCCESS" {
		logger.SugarLog.Errorf("Jenkins project %s build failed,the task_id %s", p.JobName, p.TaskID)
		UpdateXmlTaskStatus(p.TaskID, "failed", err)

		jenkinsAction := fmt.Sprintf("Jenkins project %s  build failed", p.JobName)
		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(p.TaskID, "0", nil, &jenkinsAction)
		if callErr != nil {
			logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,jenkinsAction:%s", p.TaskID, jenkinsAction)
		}
		return err
	}

	//2、调用jenkins(iBizXML-通过源文件自动获取目标路径)
	xmlPaths, err := getXMLPathByBugNumber(*p.Common)
	if err != nil {
		logger.SugarLog.Errorf("getXMLPathByBugNumber failed,err:%v", err)
		UpdateXmlTaskStatus(p.TaskID, "failed", err)

		jenkinsAction := fmt.Sprintf("getXMLPathByBugNumber failed,common id:%s", *p.Common)
		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(p.TaskID, "0", nil, &jenkinsAction)
		if callErr != nil {
			logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,jenkinsAction:%s", p.TaskID, jenkinsAction)
		}
		return err
	}

	for _, xmlPath := range xmlPaths {
		var relativePath string
		index := strings.Index(xmlPath, "iBizXML")
		if index != -1 {
			relativePath = xmlPath[index:]
		}
		param := paramBuildXML{Host: p.Host, SrcFile: xmlPath, DestPath: relativePath}

		paramMap, err := pack.StructToMap(&param)
		if err != nil {
			logger.SugarLog.Errorf("Jenkins project %s build failed,taskID:%s,host:%s,common:%s", p.JobName, p.TaskID, p.Host, p.Common)
			failedTasks = append(failedTasks, relativePath)
			continue
		}
		respBuild, err := pack.JenkinsBuild(buildXMLJobName, paramMap)
		if err != nil || respBuild.BuildResult != "SUCCESS" {
			logger.SugarLog.Errorf("Jenkins project %s build failed,the task_id %s", p.JobName, p.TaskID)
			failedTasks = append(failedTasks, relativePath)
			continue
		}
	}

	//3、写进表中并通知第三方接口

	if len(failedTasks) > 0 {
		buildStatus := "0"
		jenkinsAction := "更新xml失败"
		status := 2

		if err := mysql.InsertXml(p, status); err != nil {
			logger.SugarLog.Errorf("mysql.InsertXml failed,task_id:%s,err:%v", p.TaskID, err)
		}
		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(p.TaskID, buildStatus, nil, &jenkinsAction)
		if callErr != nil {
			logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,buildStatus:%s,jenkinsAction:%s", p.TaskID, buildStatus, jenkinsAction)
		}
		UpdateXmlTaskStatus(p.TaskID, "failed", err)
		return fmt.Errorf("task failed,taskID:%s", p.TaskID)
	} else {
		buildStatus := "1"
		jenkinsAction := "更新xml成功"
		status := 1

		if err := mysql.InsertXml(p, status); err != nil {
			logger.SugarLog.Errorf("mysql.InsertXml failed,task_id:%s,err:%v", p.TaskID, err)
		}
		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(p.TaskID, buildStatus, nil, &jenkinsAction)
		if callErr != nil {
			logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,buildStatus:%s,jenkinsAction:%s", p.TaskID, buildStatus, jenkinsAction)
		}
		UpdateXmlTaskStatus(p.TaskID, "success", nil)
		return nil
	}

}
