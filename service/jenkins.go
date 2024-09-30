package service

import (
	"context"
	"errors"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"strings"
	"time"

	"github.com/bndr/gojenkins"
)

var (
	jenkinsUrl              = "http://10.0.0.130:8080/jenkins"
	username         string = "lixiaolin"
	password         string = "Sbihss46589"
	ErrorInvalidType        = errors.New("无效的参数类型")
	newJenkinsJobStr        = "easysellers"
)

func JenkinsBuild(params map[string]string) (data *model.RespBuild, err error) {
	ctx := context.Background()

	//判断使用哪个jenkins链接信息

	jobName := params["job_name"]
	if strings.Contains(jobName, newJenkinsJobStr) {
		username = "admin"
		password = "admin@123"
		jenkinsUrl = "http://jenkins.tongtool.com"
	}

	jenkins := gojenkins.CreateJenkins(nil, jenkinsUrl, username, password)
	_, err = jenkins.Init(ctx)

	if err != nil {
		logger.SugarLog.Errorf("jenkins init failed,err:%v", err)
		return
	}

	//构建job

	queueID, err := jenkins.BuildJob(ctx, params["job_name"], params)

	if err != nil {
		logger.SugarLog.Errorf("jenkins BuildJob failed,err:%v", err)
		return
	}

	build, err := jenkins.GetBuildFromQueueID(ctx, queueID)
	if err != nil {
		logger.SugarLog.Errorf("jenkins GetQueueItem failed,err:%v", err)
		return
	}
	for build.IsRunning(ctx) {
		time.Sleep(5000 * time.Millisecond)
		build.Poll(ctx)
	}

	data = &model.RespBuild{
		BuildNumber: build.GetBuildNumber(),
		Result:      build.GetResult(),
	}

	return

}
