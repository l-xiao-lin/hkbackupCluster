package cronjobs

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/callThirdPartyAPI"
	"hkbackupCluster/pkg/pack"
	"hkbackupCluster/pkg/sendEmail"
	"hkbackupCluster/service"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	jobName               = "pipeline-ERP-主备机间隔部署"
	standalonePattern     = regexp.MustCompile(`^standalone:[^:]+:[^:]+:sdk`)
	allEnvHostsString     = "standalone:guanwang:guanwang-i2:sdk:!108ZhiYuan:!56MengNuo"
	excludeZhiYuanAllHost = "standalone:guanwang:guanwang-i2:sdk:!56MengNuo"
	changeServiceStatus   = "修改jboss和SDK状态"
	canaryJenkinsJobName  = "erp2.0美西SDK_通过主机列表_灰度调度"
)

type Text struct {
	Content             string   `json:"content"`
	MentionedMobileList []string `json:"mentioned_mobile_list"`
}

type WeChatBot struct {
	MsgType string `json:"msgtype"`
	Text    Text   `json:"text"`
}

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

func SentWeChatBot(param WeChatBot) (err error) {
	payload, err := json.Marshal(param)
	if err != nil {
		logger.SugarLog.Errorf("json.Marshal failed,err:%v", err)
		return
	}
	apiURL := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=6f39b7e2-96cb-4068-bf7e-85a573bdef38")
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		logger.SugarLog.Errorf("http.Post failed,err:%v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}
	return nil
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
	var notifiedDBColleague bool
	var triggerSQLSuccess bool
	triggerMaxRetries := 3
	for _, record := range records {
		triggerSQLSuccess = false
		//停服处理
		switch record.Status {
		case 0:
			if record.IsSqlExec {
				shouldReturn = true
				var host string
				if standalonePattern.MatchString(record.Host) {
					host = record.Host + ":!108ZhiYuan:!56MengNuo"
				} else {
					host = record.Host
				}

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
				if resp != nil && resp.BuildResult == "SUCCESS" {
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

				//四、微信及邮件通知DB同事(仅通知一次)
				if !notifiedDBColleague {
					p := WeChatBot{
						MsgType: "text",
						Text: Text{
							Content:             fmt.Sprintf("有sql语句需要执行"),
							MentionedMobileList: []string{"all"},
						},
					}
					//发送企业微信机器人
					if err := SentWeChatBot(p); err != nil {
						logger.SugarLog.Errorf("SentWeChatBot failed,err:%v", err)
					} else {
						logger.SugarLog.Infof("SentWeChatBot success,the taskID:%s", record.TaskID)
					}

					//邮件通知
					if err := sendEmail.SendEmail(); err != nil {
						logger.SugarLog.Errorf("SendEmail failed,err:%v", err)
					}
					notifiedDBColleague = true
				}

				//五、调用第三方接口触发执行SQL语句

				for attempt := 1; attempt < triggerMaxRetries; attempt++ {
					err := callThirdPartyAPI.TriggerSQLExecution(record.TaskID)
					if err == nil {
						triggerSQLSuccess = true
						break
					}
					logger.SugarLog.Errorf("Attempt %d failed: TriggerSQLExecution failed,taskID:%s,err:%v", attempt, record.TaskID, err)

					time.Sleep(time.Second * 30)
				}

				if !triggerSQLSuccess {
					//微信通知有脚本执行失败
					param := service.WeChatMessageErp{
						Message: fmt.Sprintf("TriggerSQLExecution failed,taskID:%s,err:%v", record.TaskID, err),
						CorpID:  "wxe7c550bbbe301cd3",
						Secret:  "UrUJW6Fmgdbg3vFVmssOZ6UhIThmetQeqhfmTjMVSGs",
						ToParty: "8",
						AgentID: 1000005,
					}
					service.SendWeChatAlert(param.Message, param.CorpID, param.Secret, param.ToParty, param.AgentID)

					continue
				}

				logger.SugarLog.Infof("TriggerSQLExecution success.")

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

	//有规则包的先处理
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

	//再处理没有规则包
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

	//5、针对已执行停服的主机恢复监控配置及灰度处理

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

			if resp != nil && resp.BuildResult == "SUCCESS" {
				logger.SugarLog.Infof("resume monitor success,taskID:%s,host:%s", record.TaskID, record.Host)
			} else {
				logger.SugarLog.Errorf("resume monitor failed,taskID:%s,host:%s", record.TaskID, record.Host)
			}

		}
		//灰度处理
		paramMap := make(map[string]string)
		if record.CanaryStatus != nil {
			switch *record.CanaryStatus {
			case 1:
				paramMap = map[string]string{
					"Action":       "canary",
					"canary_hosts": record.Host,
				}
			case 2:
				paramMap = map[string]string{
					"Action": "nocanary",
				}
			default:
				logger.SugarLog.Infof("Unknown CanaryStatus value %d for taskID %s, skipping...", *record.CanaryStatus, record.TaskID)
				continue
			}

			resp, buildErr := pack.JenkinsBuild(canaryJenkinsJobName, paramMap)

			if buildErr != nil {
				logger.SugarLog.Errorf("Jenkins project %s build failed for taskID %s,host %s,err:%v", canaryJenkinsJobName, record.TaskID, record.Host, buildErr)
				continue
			}

			if resp != nil && resp.BuildResult == "SUCCESS" {
				logger.SugarLog.Infof("Jenkins project %s build success,param:%v", canaryJenkinsJobName, paramMap)
			} else {
				logger.SugarLog.Errorf("Jenkins project %s build failed,taskID:%s,err:%v", canaryJenkinsJobName, record.TaskID, buildErr)
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
			now := time.Now()
			if now.Hour() == 12 && now.Minute() >= 0 && now.Minute() < 60 {
				return excludeZhiYuanAllHost
			} else {
				return allEnvHostsString
			}
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
	if resp != nil && resp.BuildResult == "SUCCESS" && buildErr == nil {
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
