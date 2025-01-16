package cronjobs

import (
	"errors"
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	jobName             = "pipeline-ERP-主备机间隔部署"
	standalonePattern   = regexp.MustCompile(`^standalone:[^:]+:[^:]+:sdk`)
	allEnvHostsString   = "standalone:guanwang:guanwang-i2:!108ZhiYuan:!56MengNuo"
	changeServiceStatus = "修改jboss和SDK状态"
)

func updateStatusBasedOnResult(taskID string, resp *pack.RespBuild, buildErr error) error {
	var status int8
	if errors.Is(buildErr, pack.ErrorConnFailed) {
		status = 2
		return mysql.UpdateReleaseStatus(taskID, status, nil)
	} else if buildErr != nil {
		status = 2
		if resp == nil {
			return mysql.UpdateReleaseStatus(taskID, status, nil)
		} else {
			return mysql.UpdateReleaseStatus(taskID, status, &resp.BuildNumber)
		}
	}

	if resp != nil {
		if resp.BuildResult == "SUCCESS" {
			status = 1
			return mysql.UpdateReleaseStatus(taskID, status, &resp.BuildNumber)
		} else {
			status = 2
			return mysql.UpdateReleaseStatus(taskID, status, &resp.BuildNumber)
		}
	}
	return nil
}

func DemoReleaseHandler(c *gin.Context) {
	err := RunReleaseTask()
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

func RunReleaseTask() (err error) {
	//1、查询package_operations中待发布的任务
	logger.SugarLog.Infof("begin RunReleaseTask func at %s", time.Now().Format("2006-01-02 15:04:05"))
	records, err := mysql.GetUnreleasedRecords()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		logger.SugarLog.Infof("no release records")
		return nil
	}

	//2、检测是否需要停服执行sql
	var shouldReturn bool
	for _, record := range records {
		//停服处理
		switch record.Status {
		case 0:
			if record.IsSqlExec {
				shouldReturn = true
				var host string
				if standalonePattern.MatchString(record.Host) {
					host = record.Host + "!108ZhiYuan:!56MengNuo"
				}
				host = record.Host

				param := &model.ServiceStop{
					Host:               host,
					RemoveJbossMonitor: true,
					RemoveSdkMonitor:   true,
					StopJboss:          true,
					StopSdk:            true,
				}
				//一、调用停服jenkins工程
				paramMap, err := pack.StructToMap(param)
				if err != nil {
					logger.SugarLog.Errorf("pack.StructToMap failed,Cannot proceed with Service stop,taskID:%s,host:%s", record.TaskID, record.Host)
					continue
				}

				logger.SugarLog.Infof("Begin stop Service,taskID:%s", record.TaskID)
				resp, buildErr := pack.JenkinsBuild(changeServiceStatus, paramMap)
				if buildErr != nil {
					logger.SugarLog.Errorf("stop Service failed,taskID:%s,host:%s", record.TaskID, record.Host)
					continue
				}

				//二、将status状态改成 已停服待执行sql脚本,以防止下次被匹配到
				var status int8
				if resp.BuildResult == "SUCCESS" {
					status = 3
				}

				if err := mysql.UpdateReleaseStatus(record.TaskID, status, nil); err != nil {
					logger.SugarLog.Errorf("mysql.UpdateReleaseStatus faied,taskID:%s,host:%s", record.TaskID, record.Host)
					continue
				}

				//三、通知第三方接口
				notificationStatus := "3"
				var jenkinsAction string

				callErr := callThirdPartyAPI.JenkinsBuildResultRsync(record.TaskID, notificationStatus, buildErr, &jenkinsAction)

				if callErr != nil {
					logger.SugarLog.Errorf("callThirdPartyAPI.JenkinsBuildResultRsync failed,err:%v", callErr)
				}
			}
		case 3:
			//已停服待执行sql
			shouldReturn = true
		}
	}
	if shouldReturn {
		return nil
	}

	//3、将release_operations status改成-1
	var status int8 = -1
	for _, record := range records {
		err = mysql.UpdateReleaseStatus(record.TaskID, status, nil)
		if err != nil {
			logger.SugarLog.Errorf("mysql.UpdateReleaseStatus failed at %d:%v", record.TaskID, err)
			continue
		}
	}

	//4、构造ansible hosts名 并初始化发版参数
	var (
		hostForRulePackage0    []string
		hostForRulePackage1    []string
		taskIdsForRulePackage0 []string
		taskIdsForRulePackage1 []string
		pkgNameForRulePackage1 string
	)
	for _, record := range records {
		if record.RmRulePackage == false {
			hostForRulePackage0 = append(hostForRulePackage0, record.Host)
			taskIdsForRulePackage0 = append(taskIdsForRulePackage0, record.TaskID)

		} else {
			hostForRulePackage1 = append(hostForRulePackage1, record.Host)
			taskIdsForRulePackage1 = append(taskIdsForRulePackage1, record.TaskID)
			pkgNameForRulePackage1 = *record.PkgName
		}
	}

	if len(hostForRulePackage0) > 0 {
		hostStringForRulePackage0 := deDuplicateHosts(hostForRulePackage0)
		paramForRulePackage0 := mysql.ReleaseOperation{
			Host:          hostStringForRulePackage0,
			RmRulePackage: false,
			PkgName:       nil,
		}
		if err = releaseUpdateAndNotify(&paramForRulePackage0, taskIdsForRulePackage0); err != nil {
			return err
		}
	}

	if len(hostForRulePackage1) > 0 {
		hostStringForRulePackage1 := deDuplicateHosts(hostForRulePackage1)
		paramForRulePackage1 := mysql.ReleaseOperation{
			Host:          hostStringForRulePackage1,
			RmRulePackage: true,
			PkgName:       &pkgNameForRulePackage1,
		}
		if err = releaseUpdateAndNotify(&paramForRulePackage1, taskIdsForRulePackage1); err != nil {
			return err
		}
	}

	//5、针对已执行停服的主机恢复监控配置

	for _, record := range records {
		if record.IsSqlExec {
			param := &model.ServiceStop{Host: record.Host, RestartJboss: false, RestartSdk: false, KeepMonitorSdk: true,
				KeepMonitorJboss: true, RemoveJbossMonitor: false, RemoveSdkMonitor: false, StopJboss: false, StopSdk: false}
			paramMap, err := pack.StructToMap(param)
			if err != nil {
				logger.SugarLog.Errorf(" pack.StructToMap failed,param host:%s,err:%v", record.Host, err)
				continue
			}
			logger.SugarLog.Infof("Begin resume monitor,taskID:%s", record.TaskID)
			resp, buildErr := pack.JenkinsBuild(changeServiceStatus, paramMap)
			if buildErr != nil {
				logger.SugarLog.Errorf("resume monitor failed,taskID:%s,host:%s", record.TaskID, record.Host)
				continue
			}

			if resp.BuildResult == "SUCCESS" {
				logger.SugarLog.Infof("resume monitor success,taskID:%s,host:%s", record.TaskID, record.Host)
			} else {
				logger.SugarLog.Errorf("resume monitor failed,taskID:%s,host:%s", record.TaskID, record.Host)
			}

		}
	}

	return
}

func deDuplicateHosts(hosts []string) string {
	uniqueElements := make(map[string]struct{})
	var allParts []string
	for _, host := range hosts {
		if standalonePattern.MatchString(host) {
			return allEnvHostsString
		}
		parts := strings.Split(host, ":")
		allParts = append(allParts, parts...)
	}
	//去重处理
	var result []string
	for _, part := range allParts {
		if _, exists := uniqueElements[part]; !exists {
			uniqueElements[part] = struct{}{}
			result = append(result, part)
		}
	}
	return strings.Join(result, ":")
}

func releaseUpdateAndNotify(param *mysql.ReleaseOperation, taskIDs []string) (err error) {
	paramMap, err := pack.StructToMap(param)
	if err != nil {
		logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
		return fmt.Errorf("pack.StructToMap failed,err:%v", err)
	}
	//调用jenkins发版
	logger.SugarLog.Infof("Begin relese processRecord")
	resp, buildErr := pack.JenkinsBuild(jobName, paramMap)

	var buildStatus string
	var jenkinsAction string
	if resp.BuildResult == "SUCCESS" && buildErr == nil {
		buildStatus = "1" //发版成功
	} else {
		buildStatus = "0" //发版失败
		jenkinsAction = "发版失败"
	}

	//更新状态
	for _, taskID := range taskIDs {
		updateErr := updateStatusBasedOnResult(taskID, resp, buildErr)

		callErr := callThirdPartyAPI.JenkinsBuildResultRsync(taskID, buildStatus, buildErr, &jenkinsAction)

		if callErr != nil {
			logger.SugarLog.Errorf("callThirdPartyAPI.JenkinsBuildResultRsync failed,err:%v", callErr)
		}
		if updateErr != nil {
			logger.SugarLog.Errorf("updateStatusBasedOnResult failed,err:%v", updateErr)
			continue
		}
	}
	return

}
