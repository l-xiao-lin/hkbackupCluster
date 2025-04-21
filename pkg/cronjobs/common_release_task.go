package cronjobs

import (
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var jobNameMap = map[string]string{
	"member":   "member版本发布",
	"passport": "passport版本发布",
	"openapi":  "Openapi版本发布",
	"tongtool": "tongtool版本发布",
	"logi":     "logi版本发布",
	"mt":       "mt版本发布",
}

type paraCommon struct {
	Host         string `json:"host"`
	SourcePath   string `json:"source_path"`
	BackupCode   bool   `json:"backup_code"`
	UploadConfig bool   `json:"upload_config"`
	OpenSchema   bool   `json:"open_schema"`
}

var serviceParamsMap = map[string]paraCommon{
	"member": {
		Host:         "member-1:member-2",
		SourcePath:   "/home/tomcat/ansible/src/Common/member",
		BackupCode:   true,
		UploadConfig: false,
		OpenSchema:   true,
	},
	"passport": {
		Host:         "passport-1:passport-2",
		SourcePath:   "/home/tomcat/ansible/src/Common/passport",
		BackupCode:   true,
		UploadConfig: false,
	},
	"tongtool": {
		Host:         "tongtool-1:tongtool-2",
		SourcePath:   "/home/tomcat/ansible/src/Common/tongtool",
		BackupCode:   true,
		UploadConfig: false,
	},
	"logi": {
		Host:         "logi-2:logi-1",
		SourcePath:   "/home/tomcat/ansible/src/Common/logi",
		BackupCode:   true,
		UploadConfig: false,
		OpenSchema:   false,
	},
	"Openapi": {
		Host:         "openapi-3:openapi-2:openapi-1",
		SourcePath:   "/home/tomcat/ansible/src/Common/openapi",
		BackupCode:   true,
		UploadConfig: false,
		OpenSchema:   false,
	},
	"mt": {
		Host:         "mt",
		SourcePath:   "/home/tomcat/ansible/src/Common/mt",
		BackupCode:   true,
		UploadConfig: false,
		OpenSchema:   false,
	},
}

func DemoCommonReleaseHandler(c *gin.Context) {
	err := RunCommonReleaseTask()
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

func RunCommonReleaseTask() (err error) {
	//1、查询common_release_operations中待发布的任务
	logger.SugarLog.Infof("begin RunCommonReleaseTask func at %s", time.Now().Format("2006-01-02 15:04:05"))
	records, err := mysql.GetCommonUnreleasedRecords()
	if err != nil {
		return
	}
	if len(records) == 0 {
		logger.SugarLog.Infof("RunCommonReleaseTask no release records")
		return
	}

	//2、将满足条件记录的状态改成-1
	var status = -1
	for _, record := range records {
		err = mysql.UpdateCommonReleaseStatus(record.TaskID, status, nil)
		if err != nil {
			logger.SugarLog.Errorf("ysql.UpdatePackCommonStatus taskID:%s,err:%v", record.TaskID, err)
			continue
		}
	}

	//3、构建参数调用发版工程
	for _, record := range records {

		commonJobName, exist := jobNameMap[record.ServiceName]
		if !exist {
			logger.SugarLog.Errorf("unknown service name: %s", record.ServiceName)
			continue
		}

		param, exist := serviceParamsMap[record.ServiceName]
		if !exist {
			logger.SugarLog.Errorf("unknow  serviceParamsMap service name:%s", record.ServiceName)
			continue
		}

		if record.OpenSchema {
			param.OpenSchema = true
		}

		if record.HasConfiguration {
			param.UploadConfig = true
		}

		paramMap, err := pack.StructToMap(&param)
		if err != nil {
			logger.SugarLog.Errorf("pack.StructToMap faied,param:%+v", param)
			continue
		}

		resp, buildErr := pack.JenkinsBuild(commonJobName, paramMap)

		var buildStatus string
		var jenkinsAction string
		if resp != nil && resp.BuildResult == "SUCCESS" && buildErr == nil {
			buildStatus = "1" //发版成功
		} else {
			buildStatus = "0" //发版失败
			jenkinsAction = "发版失败"
		}

		updateErr := updateCommonStatusBasedOnResult(record.TaskID, resp, buildErr)

		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, buildStatus, buildErr, &jenkinsAction)

		if callErr != nil {
			logger.SugarLog.Errorf("callThirdPartyAPI.JenkinsBuildResultRsync failed,taskID:%s err:%v", record.TaskID, callErr)
		}
		if updateErr != nil {
			logger.SugarLog.Errorf("updateStatusBasedOnResult failed,taskID:%s,err:%v", record.TaskID, updateErr)
			continue
		}

	}

	return
}
func updateCommonStatusBasedOnResult(taskID string, resp *pack.RespBuild, buildErr error) (err error) {
	var status int
	if buildErr != nil {
		status = 2
		if resp == nil {
			return mysql.UpdateCommonReleaseStatus(taskID, status, nil)
		} else {
			return mysql.UpdateCommonReleaseStatus(taskID, status, &resp.BuildNumber)
		}
	}

	if resp != nil {
		if resp.BuildResult == "SUCCESS" {
			status = 1
			return mysql.UpdateCommonReleaseStatus(taskID, status, &resp.BuildNumber)
		} else {
			status = 2
			return mysql.UpdateCommonReleaseStatus(taskID, status, &resp.BuildNumber)
		}
	}
	return

}
