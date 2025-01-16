package cronjobs

import (
	"errors"
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"hkbackupCluster/pkg/sshclient"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

var (
	GetConfigJobName    = "拉取-配置文件-到堡垒机"
	ModifyConfigJobName = "新增或修改ERP2.0-部分环境配置"
	ansibleHost         = "172.16.60.1"
)

func DemoPackageHandler(c *gin.Context) {

	err := RunPackageTask()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"msg":  err.Error(),
			"code": 1002,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"msg":  "package success",
		"code": 1000,
	})

}

func updateStatusOnPackageResult(resp *pack.RespBuild, buildErr error, record *model.RespPackageData) error {
	var status int8
	if errors.Is(buildErr, pack.ErrorConnFailed) {
		//jenkins连接失败，只更新status
		status = 2
		if err := mysql.UpdatePackStatus(record.TaskID, status, nil); err != nil {
			logger.SugarLog.Errorf("mysql.UpdatePackStatus %s failed,err:%v", record.TaskID, err)
			return err
		}
	} else if buildErr != nil {
		status = 2
		if resp == nil {
			if err := mysql.UpdatePackStatus(record.TaskID, status, nil); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackStatus %s failed,err:%v", record.TaskID, err)
				return err
			}
		} else {
			if err := mysql.UpdatePackStatus(record.TaskID, status, &resp.BuildNumber); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackStatus %s failed,err:%v", record.TaskID, err)
				return err
			}
		}

	}
	if resp != nil {
		if resp.BuildResult == "SUCCESS" {
			status = 1
			if err := mysql.HandleSuccessPackTransaction(record.TaskID, status, &resp.BuildNumber, record); err != nil {
				logger.SugarLog.Errorf("mysql.HandleSuccessPackTransaction %s failed,err:%v", record.TaskID, err)
				return err
			}
		} else {
			status = 2
			if err := mysql.UpdatePackStatus(record.TaskID, status, &resp.BuildNumber); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackStatus %s failed,err:%v", record.TaskID, err)
				return err

			}
		}
	}
	return nil
}

func isConfigExists(config model.Config) (bool, error) {
	client, err := sshclient.SshConnect(ansibleHost)
	if err != nil {
		return false, err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return false, err
	}
	defer session.Close()
	var filePath = fmt.Sprintf("/home/tomcat/ansible/configfiles/jboss/%s-\\$TEMPLATE\\$.properties", config.Host)
	if strings.Contains(config.ConfigContent, "sdk") {
		filePath = fmt.Sprintf("/home/tomcat/ansible/configfiles/sdk/%s-ibiz_sdk.properties", config.Host)
	}
	cmd := fmt.Sprintf("grep -qF %s %s; echo $?", config.ConfigContent, filePath)
	output, err := session.CombinedOutput(cmd)
	if err == nil && len(output) > 0 {
		exitStatus := strings.TrimSpace(string(output))
		if exitStatus == "0" {
			return true, nil
		} else if exitStatus == "1" {
			return false, nil
		}
	}
	return false, nil
}

func RunPackageTask() (err error) {
	//1、从package_operations中查询满足条件的记录
	zap.L().Info("Begin RunPackageTask ", zap.String("time", time.Now().Format("2006-01-02 15:04:05")))
	records, err := mysql.GetUnPackageRecords()
	if err != nil {
		return
	}
	if len(records) == 0 {
		logger.SugarLog.Infof("no release records")
		return fmt.Errorf("no release records")
	}

	//2、将packaging_time status 改成-1
	var status int8 = -1
	for _, record := range records {
		err = mysql.UpdatePackStatus(record.TaskID, status, nil)
		if err != nil {
			logger.SugarLog.Errorf("mysql.UpdatePackStatus failed,err:%v", err)
			continue
		}
	}

	//3、检测是否需要更新配置文件
	processedHost := make(map[string]bool)
	for _, record := range records {
		if record.UpdateJbossConf || record.UpdateSdkConf {
			configurations, err := mysql.GetConfigurations(record.TaskID)
			if err != nil {
				logger.SugarLog.Errorf("mysql.GetConfigurations failed,taskID:%s,err:%v", record.TaskID, err)
				continue
			}

			//一、拉取配置文件仅当主机未被处理时才执行
			if _, exists := processedHost[record.Host]; !exists {
				resp, err := pack.JenkinsBuild(GetConfigJobName, map[string]string{"host": record.Host})
				if err != nil || resp.BuildResult != "SUCCESS" {
					logger.SugarLog.Errorf("pack.JenkinsBuild GetConfigJobName failed,host:%s,err:%v", record.Host, err)
					continue
				}
				logger.SugarLog.Infof("pack.JenkinsBuild GetConfigJobName success,host:%s", record.Host)
				processedHost[record.Host] = true

			}

			//二、循环修改配置文件

			for _, configuration := range configurations {
				//判断新增的配置文件是否已存在
				isExist, err := isConfigExists(configuration)
				if err != nil {
					logger.SugarLog.Errorf("isConfigExists config:%v ,err:%v", configuration, err)
					continue
				}

				if !isExist {
					paramMap, err := pack.StructToMap(&configuration)
					if err != nil {
						logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
						continue
					}
					resp, err := pack.JenkinsBuild(ModifyConfigJobName, paramMap)
					if err != nil || resp.BuildResult != "SUCCESS" {
						logger.SugarLog.Errorf("pack.JenkinsBuild ModifyConfigJobName failed,host:%s,err:%v", record.Host, err)
						continue
					}
					logger.SugarLog.Infof("pack.JenkinsBuild ModifyConfigJobName success.")
				}

			}

		}
	}

	//4、循环遍历开始调用jenkins打包工程

	var buildStatus string
	for _, record := range records {
		//判断是否有包，如果没有则无需打包
		if record.IsPackage {
			paramMap, err := pack.StructToMap(&record)
			if err != nil {
				logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
				continue
			}
			resp, buildErr := pack.JenkinsBuild(record.JobName, paramMap)

			var jenkinsAction string
			if resp.BuildResult == "SUCCESS" {
				buildStatus = "2" //给第三方接口返回 已打包
			} else {
				buildStatus = "0" //给第三方接口返回 退回
				jenkinsAction = "打包失败"
			}

			updateErr := updateStatusOnPackageResult(resp, buildErr, &record)

			callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, buildStatus, buildErr, &jenkinsAction)

			if callErr != nil {
				logger.SugarLog.Errorf("callThirdPartyAPI.JenkinsBuildResultRsync failed,err:%v", callErr)
			}

			if updateErr != nil {
				logger.SugarLog.Errorf("updateStatusOnPackageResult failed,err:%v", updateErr)
				continue
			}
		} else {
			status = 1 //无需要打包，更新状态并写到release_operations表中
			if err := mysql.HandleSuccessPackTransaction(record.TaskID, status, nil, &record); err != nil {
				logger.SugarLog.Errorf("mysql.HandleSuccessPackTransaction taskID:%s err:%v", record.TaskID, err)
				continue
			}

		}
	}
	return nil

}
