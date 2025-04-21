package cronjobs

import (
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
	GetCommonConfJobName    = "C-0拉取公共服务配置"
	UpdateCommonConfJobName = "C-1修改公共服务配置"
	commonPackName          = "公共服务上传并自动打包"
	getVersionHost          = "10.0.0.130"
	maxRetries              = 3
)

type CommonConfigCheckResult struct {
	Status      string
	Description string
	Error       error
}

func DemoCommonPackageHandler(c *gin.Context) {

	err := RunCommonPackageTask()
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

func RunCommonPackageTask() (err error) {
	//1、从common_package_operations查询出满足条件的记录
	zap.L().Info("Begin RunCommonPackageTask ", zap.String("time", time.Now().Format("2006-01-02 15:04:05")))
	records, err := mysql.GetUnPackageCommonRecords()
	if err != nil {
		return
	}
	if len(records) == 0 {
		logger.SugarLog.Infof("RunCommonPackageTask no release records")
		return
	}
	//2、将status状态改成-1
	var status = -1
	for _, record := range records {
		fmt.Printf("package_time:%s", record.PackageTime)
		err = mysql.UpdatePackCommonStatus(record.TaskID, status, nil)
		if err != nil {
			logger.SugarLog.Errorf("mysql.UpdatePackCommonStatus failed,taskID:%s err:%v", record.TaskID, err)
			continue
		}
	}

	//3、检测是否存在配置文件
	processedCommonServiceName := make(map[string]bool)
	configurationUpdateResults := make(map[string]bool)
	for _, record := range records {
		configUpdateSuccess := true
		if record.HasConfiguration {
			configurations, err := mysql.GetCommonConfigurations(record.TaskID)
			if err != nil {
				logger.SugarLog.Errorf("mysql.GetCommonConfigurations failed,taskID:%s,err:%v", record.TaskID, err)
				configUpdateSuccess = false
			} else {
				for _, configuration := range configurations {
					//拉取指定公共服务配置,同一个项目只拉取一次
					if _, exists := processedCommonServiceName[configuration.ServiceName]; !exists {
						processedCommonServiceName[configuration.ServiceName] = true
						resp, err := pack.JenkinsBuild(GetCommonConfJobName, map[string]string{"common_service": configuration.ServiceName})
						if err != nil || resp.BuildResult != "SUCCESS" {
							logger.SugarLog.Errorf("pack.JenkinsBuild failed,jenkins jobName:%s taskID:%s", GetCommonConfJobName, configuration.TaskID)
							configUpdateSuccess = false
							break
						}
					}

					//检查配置文件状态
					checkResult, err := checkCommonConfiguration(configuration)
					if err != nil {
						logger.SugarLog.Errorf("checkCommonConfiguration failed,config:%s,err:%v", configuration.Content, err)
						configUpdateSuccess = false
						break
					}
					logger.SugarLog.Infof("checkCommonConfiguration config:%s checkResult:%#v", configuration.Content, checkResult)
					for serviceName, result := range checkResult {
						if result.Status == "config_missing" {
							//修改指定公共服务配置
							paraMap := map[string]string{
								"Action":         configuration.Action,
								"Common_Project": serviceName,
								"Config_Content": configuration.Content,
							}
							updateResp, err := pack.JenkinsBuild(UpdateCommonConfJobName, paraMap)
							if err != nil || updateResp.BuildResult != "SUCCESS" {
								logger.SugarLog.Errorf("pack.JenkinsBuild failed,jenkins jobName:%s taskID:%s", UpdateCommonConfJobName, configuration.TaskID)
								configUpdateSuccess = false
								break
							}
						}
					}
				}
			}
			configurationUpdateResults[record.TaskID] = configUpdateSuccess
			//配置更新失败
			if !configUpdateSuccess {
				status = 2
				buildStatus := "0" //给第三方接口返回 退回
				jenkinsAction := fmt.Sprintf("%s公共服务更新配置失败", record.ServiceName)

				updateErr := mysql.UpdatePackCommonStatus(record.TaskID, status, nil)
				if updateErr != nil {
					logger.SugarLog.Errorf("mysql.UpdatePackCommonStatus failed,taskID:%s serverName:%s", record.TaskID, record.ServiceName)
				}
				callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, buildStatus, nil, &jenkinsAction)
				if callErr != nil {
					logger.SugarLog.Errorf("JenkinsBuildResultRsync failed,taskID:%s,buildStatus:%s,jenkinsAction:%s", record.TaskID, buildStatus, jenkinsAction)
				}
			}
		}
	}

	//4、调用jenkins打包工程
	for _, record := range records {
		if record.HasConfiguration {
			if configUpdateSuccess, ok := configurationUpdateResults[record.TaskID]; !ok || !configUpdateSuccess {
				logger.SugarLog.Errorf("Configuration update failed,taskID:%s", record.TaskID)
				continue //如果存在配置文件，并且配置文件更新失败，则直接执行下一条任务
			}
		}
		//开始打公共服务的包
		if err := handlerCommonPackage(record); err != nil {
			logger.SugarLog.Errorf("Failed to package for taskID:%s,err:%v", record.TaskID, err)
			continue
		}
	}
	return
}

func getCommonPackVersionWithRetry(host, cmd string, maxRetries int) (string, error) {

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		version, err := getCommonPackVersion(host, cmd)
		if err == nil {
			return version, nil //如果成功，直接返回
		}
		lastErr = err

		logger.SugarLog.Warnf("Attemp %d failed for getCommonPackVersionWithRetry,Retrying...Error:%v", i, err)
		time.Sleep(time.Second * 30)
	}
	return "", fmt.Errorf("all %d attemps failed for getCommonPackVersionWithRetry. Last err:%v", maxRetries, lastErr)

}

func getCommonPackVersion(host, cmd string) (string, error) {

	client, err := sshclient.SshConnect(host)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	cmdInfo, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(cmdInfo)), nil

}

func handlerCommonPackage(record model.RespPackageCommonData) (err error) {
	//获取version参数
	cmd := fmt.Sprintf(" ls -l /data/erp2_deploy/Common/%s | awk 'NR >1{print $NF}'|sort -nr|head -1", record.ServiceName)
	version, err := getCommonPackVersionWithRetry(getVersionHost, cmd, maxRetries)
	if err != nil {
		logger.SugarLog.Errorf("getCommonPackVersion failed,serviceName:%s,err:%v", record.ServiceName, err)
		return
	}

	paraMap := map[string]string{
		"common_service": record.ServiceName,
		"version":        version,
	}
	resp, buildErr := pack.JenkinsBuild(commonPackName, paraMap)
	if buildErr != nil {
		logger.SugarLog.Errorf("pack.JenkinsBuild failed,jobName:%s,serviceName:%s,err:%v", commonPackName, record.ServiceName, err)
		return
	}
	var buildStatus string
	var jenkinsAction string
	if resp != nil && resp.BuildResult == "SUCCESS" {
		buildStatus = "2" //给第三方接口返回 已打包
		jenkinsAction = fmt.Sprintf("公共服务:%s 打包成功", record.ServiceName)
	} else {
		buildStatus = "0" //给第三方接口返回 退回
		jenkinsAction = fmt.Sprintf("公共服务:%s 打包失败", record.ServiceName)

	}
	updateErr := updateStatusOnCommonPackageResult(resp, buildErr, &record)

	callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, buildStatus, buildErr, &jenkinsAction)

	if callErr != nil {
		logger.SugarLog.Errorf("callThirdPartyAPI.JenkinsBuildResultRsync failed,err:%v", callErr)
	}
	if updateErr != nil {
		logger.SugarLog.Errorf("updateStatusOnPackageResult failed,err:%v", updateErr)
	}
	return nil

}

func updateStatusOnCommonPackageResult(resp *pack.RespBuild, buildErr error, record *model.RespPackageCommonData) (err error) {
	var status int
	if buildErr != nil {
		status = 2
		if resp == nil {
			if err := mysql.UpdatePackCommonStatus(record.TaskID, status, nil); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackCommonStatus taskID:%s failed", record.TaskID)
				return err
			}
		} else {
			if err := mysql.UpdatePackCommonStatus(record.TaskID, status, &resp.BuildNumber); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackCommonStatus taskID:%s failed", record.TaskID)
				return err
			}
		}
	}
	if resp != nil {
		if resp.BuildResult == "SUCCESS" {
			status = 1
			if err := mysql.HandleSuccessCommonPackTransaction(record.TaskID, status, &resp.BuildNumber, record); err != nil {
				logger.SugarLog.Errorf("mysql.HandleSuccessCommonPackTransaction taskID:%s failed,err:%v", record.TaskID, err)
				return err
			}

		} else {
			status = 2
			if err := mysql.UpdatePackCommonStatus(record.TaskID, status, &resp.BuildNumber); err != nil {
				logger.SugarLog.Errorf("mysql.UpdatePackCommonStatus taskID:%s failed", record.TaskID)
				return err
			}

		}
	}
	return
}

func checkCommonConfiguration(config model.RespConfCommon) (map[string]CommonConfigCheckResult, error) {
	client, err := sshclient.SshConnect(ansibleHost)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	executeCommand := func(command string) (string, error) {
		session, err := client.NewSession()
		if err != nil {
			logger.SugarLog.Errorf("failed to create new session for commonServerName %s:%v", config.ServiceName, err)
			return "", err
		}
		defer session.Close()

		output, err := session.CombinedOutput(command)
		if err != nil {
			logger.SugarLog.Errorf("failed to execute command on commonServerName %s:%v", config.ServiceName, err)
			return "", err
		}
		return strings.TrimSpace(string(output)), nil
	}
	checkResult := make(map[string]CommonConfigCheckResult)
	filePath := fmt.Sprintf("/home/tomcat/ansible/common_server/config/%s/*.properties; echo $?", config.ServiceName)
	cmd := fmt.Sprintf("grep -qxF '%s' %s", config.Content, filePath)

	var configExists bool

	output, err := executeCommand(cmd)
	if err != nil {
		checkResult[config.ServiceName] = CommonConfigCheckResult{
			Status:      "Error",
			Description: "配置检测失败",
			Error:       err,
		}
	}
	configExists = output == "0"

	if configExists {
		checkResult[config.ServiceName] = CommonConfigCheckResult{
			Status:      "config_exist",
			Description: "配置已存在",
		}
	} else {
		checkResult[config.ServiceName] = CommonConfigCheckResult{
			Status:      "config_missing",
			Description: "配置缺失",
		}
	}
	return checkResult, nil
}
