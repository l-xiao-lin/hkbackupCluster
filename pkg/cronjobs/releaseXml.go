package cronjobs

import (
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"hkbackupCluster/pkg/sshclient"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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

func notifyAndUpdateXml(status int, taskID string, thirdStatus, action string, buildErr error) {

	if updateErr := mysql.UpdateXmlStatus(status, taskID); updateErr != nil {
		logger.SugarLog.Errorf(" UpdateXmlStatus failed,task_id:%s,err:%v", taskID, updateErr)
	}
	callErr := callThirdPartyAPI.JenkinsBuildResultRsync(taskID, thirdStatus, buildErr, &action)
	if callErr != nil {
		logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,jenkinsAction:%s", taskID, action)
	}

}

func DemoReleaseXmlHandler(c *gin.Context) {
	err := ReleaseXmlHandler()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1001,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "success",
		"code": 1000,
	})

}

func ReleaseXmlHandler() (err error) {
	//1、从release_xml表中查询满足条件的记录
	logger.SugarLog.Infof("begin RunReleaReleaseXmlHandlerseTask func at %s", time.Now().Format("2006-01-02 15:04:05"))
	records, err := mysql.GetUnreleasedXmlRecords()
	if err != nil {
		logger.SugarLog.Errorf("mysql.GetUnreleasedXmlRecords failed,err:%v", err)
		return err
	}

	if len(records) == 0 {
		logger.SugarLog.Infof("ReleaseXmlHandler query now records")
		return nil
	}

	//2、将状态改成进行中

	for _, record := range records {
		status := -1
		if err := mysql.UpdateXmlStatus(status, record.TaskID); err != nil {
			logger.SugarLog.Errorf("mysql.UpdateXmlStatus failed,taskID:%s,err:%s", record.TaskID, err)
			continue
		}
	}

	//3、调用jenkins(拷贝-当天增量-到堡垒机),只调用一次
	var srcPath string
	var jobNameXml string
	isExecXmlCopyProject := false //标识是否已经执行了jenkins拷贝包的工程

	for _, record := range records {
		if !isExecXmlCopyProject {
			srcPath = record.SrcPath
			jobNameXml = record.JobName
			isExecXmlCopyProject = true
			paramMap := map[string]string{
				"src_path": srcPath,
			}
			resp, err := pack.JenkinsBuild(jobNameXml, paramMap)
			if err != nil || resp.BuildResult != "SUCCESS" {
				//更新所有相关记录的状态为失败
				for _, r := range records {
					status := 2
					thirdStatus := "0"
					jenkinsAction := fmt.Sprintf("Jenkins project %s build failed", jobNameXml)
					notifyAndUpdateXml(status, r.TaskID, thirdStatus, jenkinsAction, err)
				}
				return err
			}
		}

		//获取xmlPaths
		xmlPaths, err := getXMLPathByBugNumber(*record.Common)
		if err != nil {
			status := 2
			thirdStatus := "0"
			jenkinsAction := fmt.Sprintf("getXMLPathByBugNumber failed,common id:%s", *record.Common)
			notifyAndUpdateXml(status, record.TaskID, thirdStatus, jenkinsAction, nil)
			continue
		}

		buildFailed := false
		for _, xmlPath := range xmlPaths {
			var relativePath string
			index := strings.Index(xmlPath, "iBizXML")
			if index != -1 {
				relativePath = xmlPath[index:]
			}
			param := paramBuildXML{Host: record.Host, SrcFile: xmlPath, DestPath: relativePath}

			paramMap, err := pack.StructToMap(&param)
			if err != nil {
				logger.SugarLog.Errorf("pack.StructToMap failed,taskID:%s,host:%s,common:%s", record.TaskID, record.Host, *record.Common)
				buildFailed = true
				break
			}

			//调用Jenkins buildXMLJobName工程
			respBuild, err := pack.JenkinsBuild(buildXMLJobName, paramMap)
			if err != nil || respBuild.BuildResult != "SUCCESS" {
				logger.SugarLog.Errorf("Jenkins project %s build failed,the task_id %s", buildXMLJobName, record.TaskID)
				buildFailed = true
				break
			}
		}

		if buildFailed {
			status := 2
			thirdStatus := "0"
			jenkinsAction := fmt.Sprintf("Jenkins project %s failed,common id:%s", buildXMLJobName, *record.Common)
			notifyAndUpdateXml(status, record.TaskID, thirdStatus, jenkinsAction, nil)
		} else {
			status := 1
			thirdStatus := "1"
			jenkinsAction := fmt.Sprintf("Jenkins project %s success,common id:%s", buildXMLJobName, *record.Common)
			notifyAndUpdateXml(status, record.TaskID, thirdStatus, jenkinsAction, nil)
		}
	}

	return nil
}
