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
	"hkbackupCluster/service"
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
	ansibleGroups       = map[string]bool{
		"standalone:guanwang:guanwang-i2:sdk": true,
		"monday":                              true,
		"wednesday":                           true,
		"standalone:guanwang:guanwang-i2:sdk:!monday":   true,
		"standalone:guanwang:guanwang-i2:sdk:wednesday": true,
		"sdk": true,
	}
)

type HostCheckResult struct {
	Status      string
	Description string
	Error       error
}

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

func checkConfiguration(config model.Config) (map[string]HostCheckResult, error) {
	client, err := sshclient.SshConnect(ansibleHost)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	hostList, err := generationHosts(config.Host)
	if err != nil {
		return nil, err
	}

	checkResult := make(map[string]HostCheckResult)
	for _, host := range hostList {
		var filePath = fmt.Sprintf("/home/tomcat/ansible/configfiles/jboss/%s-\\$TEMPLATE\\$.properties", host)
		if strings.Contains(config.ConfigType, "sdk") {
			filePath = fmt.Sprintf("/home/tomcat/ansible/configfiles/sdk/%s-ibiz_sdk.properties", host)
		}

		executeCommand := func(command string) (string, error) {
			session, err := client.NewSession()
			if err != nil {
				logger.SugarLog.Errorf("failed to create new session for host %s:%v", host, err)
				return "", err
			}
			defer session.Close()
			output, err := session.CombinedOutput(command)
			if err != nil {
				logger.SugarLog.Errorf("failed to execute command on host %s:%v", host, err)
				return "", err
			}
			return strings.TrimSpace(string(output)), nil
		}

		//检查文件是否存在
		fileCheckCmd := fmt.Sprintf("test -f %s;echo $?", filePath)
		output, err := executeCommand(fileCheckCmd)
		if err != nil {
			checkResult[host] = HostCheckResult{
				Status:      "Error",
				Description: "检查文件失败",
				Error:       err,
			}
			continue //执行下一个主机
		}
		fileExists := output == "0"

		var configExists bool
		if fileExists {
			content := strings.TrimSpace(config.ConfigContent)
			cmd := fmt.Sprintf("grep -qxF '%s' %s; echo $?", content, filePath)
			logger.SugarLog.Infof("cmd:%s", cmd)
			output, err := executeCommand(cmd)
			if err != nil {
				checkResult[host] = HostCheckResult{
					Status:      "Error",
					Description: "检查配置失败",
					Error:       err,
				}
				continue
			}
			configExists = output == "0"
		}

		switch {
		case !fileExists:
			checkResult[host] = HostCheckResult{
				Status:      "file_not_found",
				Description: "文件不存在",
			}
		case configExists:
			checkResult[host] = HostCheckResult{
				Status:      "config_exist",
				Description: "配置已存在",
			}
		default:
			checkResult[host] = HostCheckResult{
				Status:      "config_missing",
				Description: "配置缺失",
			}
		}

	}

	return checkResult, nil

}

func generationHosts(paramHost string) ([]string, error) {

	//如果传递进来的是ansible组名,则进行主机列表查询
	if ansibleGroups[paramHost] {
		cmd := fmt.Sprintf("ansible -i $Z_asbList_erp2 %s --list| grep -v hosts", paramHost)
		respData, err := service.GetInventory(cmd)
		if err != nil {
			zap.L().Error("service.GetInventory failed", zap.String("cmd", cmd), zap.Error(err))
			return nil, err
		}
		return respData, nil
	}
	//如果是以多个主机传递
	if strings.Contains(paramHost, ":") {
		return strings.Split(paramHost, ":"), nil
	}
	return []string{paramHost}, nil
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
		return nil
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
	configurationUpdateResults := make(map[string]bool)

	for _, record := range records {
		configUpdateSuccess := true
		if record.UpdateJbossConf || record.UpdateSdkConf {
			configurations, err := mysql.GetConfigurations(record.TaskID)
			if err != nil && configurations == nil {
				logger.SugarLog.Errorf("mysql.GetConfigurations failed,taskID:%s,err:%v", record.TaskID, err)
				configUpdateSuccess = false

			} else {
				//拉取配置文件仅当主机未被处理时才执行
				hostList, err := generationHosts(record.Host)
				if err != nil {
					logger.SugarLog.Errorf("generationHosts failed,taskID:%s host:%s", record.TaskID, record.Host)
					configUpdateSuccess = false
				} else {
					var fetchConfigurationHosts []string
					for _, host := range hostList {
						if _, exists := processedHost[host]; !exists {
							processedHost[host] = true
							fetchConfigurationHosts = append(fetchConfigurationHosts, host)
						}
					}
					if len(fetchConfigurationHosts) > 0 {
						resp, err := pack.JenkinsBuild(GetConfigJobName, map[string]string{"host": strings.Join(fetchConfigurationHosts, ":")})
						if err != nil || resp.BuildResult != "SUCCESS" {
							logger.SugarLog.Errorf("pack.JenkinsBuild GetConfigJobName failed,host:%s,err:%v", strings.Join(fetchConfigurationHosts, ":"), err)
							configUpdateSuccess = false
						}
					}
				}
				//循环修改配置文件
				for _, configuration := range configurations {
					if configUpdateSuccess {
						//检查配置文件状态
						checkResult, err := checkConfiguration(configuration)
						if err != nil {
							logger.SugarLog.Errorf("isConfigExists config:%v ,err:%v", configuration, err)
							configUpdateSuccess = false
							break
						}
						logger.SugarLog.Infof("checkConfiguration checkResult:%#v", checkResult)

						//过滤heckResult中的config_missing(需要更新的配置)及error
						var configUpdateHosts []string
						var configCheckFailedHost []string
						for host, result := range checkResult {
							if result.Status == "config_missing" {
								configUpdateHosts = append(configUpdateHosts, host)
							} else if result.Status == "Error" {
								configCheckFailedHost = append(configCheckFailedHost, host)
							}
						}
						if len(configCheckFailedHost) > 0 {
							logger.SugarLog.Errorf("exists configCheckFailedHost:%s", configCheckFailedHost)
							configUpdateSuccess = false
							break
						}

						//存在需要更新配置的主机
						if len(configUpdateHosts) > 0 {
							//判断host是否为主机组,如果是则转换成host
							host := strings.Join(configUpdateHosts, ":")
							newConfiguration := &model.Config{
								ConfigType:    configuration.ConfigType,
								ConfigContent: configuration.ConfigContent,
								ConfigAction:  configuration.ConfigAction,
								Host:          host,
							}

							paramMap, err := pack.StructToMap(newConfiguration)
							if err != nil {
								logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
								configUpdateSuccess = false
								break
							}
							resp, err := pack.JenkinsBuild(ModifyConfigJobName, paramMap)
							if err != nil || resp.BuildResult != "SUCCESS" {
								logger.SugarLog.Errorf("pack.JenkinsBuild ModifyConfigJobName failed,host:%s,err:%v", record.Host, err)
								configUpdateSuccess = false
								break
							}
							logger.SugarLog.Infof("pack.JenkinsBuild ModifyConfigJobName success.")
						}

					} else {
						break
					}
				}
			}
			configurationUpdateResults[record.TaskID] = configUpdateSuccess
			if !configUpdateSuccess {
				status = 2         //修改package_operations status为2(失败)
				buildStatus := "0" //给第三方接口返回 退回
				jenkinsAction := "更新配置失败"

				updateErr := mysql.UpdatePackStatus(record.TaskID, status, nil)
				if updateErr != nil {
					zap.L().Error("mysql.UpdatePackStatus failed", zap.String("taskID", record.TaskID), zap.Int8("status", status))
				}

				callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, buildStatus, nil, &jenkinsAction)
				if callErr != nil {
					logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,buildStatus:%s,jenkinsAction:%s", record.TaskID, buildStatus, jenkinsAction)
				}
			}
		}
	}

	//4、循环遍历开始调用jenkins打包工程

	for _, record := range records {
		//有包有配置
		if record.IsPackage && (record.UpdateJbossConf || record.UpdateSdkConf) {
			if configUpdateSuccess, ok := configurationUpdateResults[record.TaskID]; !ok || !configUpdateSuccess {
				zap.L().Error("Configuration update failed", zap.String("taskID", record.TaskID))
				continue //无需再执行后续打包
			}
			if err := handlerPackage(record); err != nil {
				logger.SugarLog.Errorf("Failed to package for taskID:%s,err:%v", record.TaskID, err)
				continue
			}
		} else if record.IsPackage { //只需要打包
			if err := handlerPackage(record); err != nil {
				logger.SugarLog.Errorf("Failed to package for taskID:%s,err:%v", record.TaskID, err)
				continue
			}
			//无包有配置
		} else if !record.IsPackage && (record.UpdateJbossConf || record.UpdateSdkConf) { //无包只有配置文件
			if configUpdateSuccess, ok := configurationUpdateResults[record.TaskID]; !ok || !configUpdateSuccess {
				logger.SugarLog.Errorf("Configuration update failed,taskID:%s,err:%v", record.TaskID, err)
				continue
			}

			record.Common = nil //将bug号改成空
			if err := handlerPackage(record); err != nil {
				logger.SugarLog.Errorf("Failed to package for taskID:%s,err:%v", record.TaskID, err)
				continue
			}

			//只需要停机跑sql,不需要打包也没有配置文件
		} else if record.IsSqlExec {
			status = 1 //将状态改成1 成功
			if err := mysql.HandleSuccessPackTransaction(record.TaskID, status, nil, &record); err != nil {
				logger.SugarLog.Errorf("HandleSuccessPackTransaction failed,taskID:%s,err:%v", record.TaskID, err)
				continue
			}
		}

	}
	return nil

}

func handlerPackage(record model.RespPackageData) error {

	paramMap, err := pack.StructToMap(&record)
	if err != nil {
		logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
		return err
	}
	resp, buildErr := pack.JenkinsBuild(record.JobName, paramMap)

	var buildStatus string
	var jenkinsAction string
	if resp != nil && resp.BuildResult == "SUCCESS" {
		buildStatus = "2" //给第三方接口返回 已打包
		jenkinsAction = "打包成功"
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
	}
	return nil
}
