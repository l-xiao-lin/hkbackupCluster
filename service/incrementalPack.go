package service

import (
	"context"
	"fmt"
	"time"

	"github.com/bndr/gojenkins"
)

var (
	Url            = "http://10.0.0.130:8080/jenkins/"
	deployUsername = "lixiaolin"
	deployPassword = "Sbihss46589"
	jobName        = "pipeline-输入bug号自动打包"
)

type RespBuild struct {
	BuildNumber int64  `json:"build_number"`
	BuildResult string `json:"build_result"`
}

func IncrementalPack(param map[string]string) (respBuild *RespBuild, err error) {

	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, Url, deployUsername, deployPassword)
	_, err = jenkins.Init(ctx)
	if err != nil {
		return nil, err
	}

	queueID, err := jenkins.BuildJob(ctx, jobName, param)
	if err != nil {
		return nil, err
	}
	build, err := jenkins.GetBuildFromQueueID(ctx, queueID)
	if err != nil {
		return nil, err
	}
	for build.IsRunning(ctx) {
		time.Sleep(5000 * time.Millisecond)
		build.Poll(ctx)
	}
	fmt.Printf("build number %d with result:%v\n", build.GetBuildNumber(), build.GetResult())
	respBuild = &RespBuild{
		BuildNumber: build.GetBuildNumber(),
		BuildResult: build.GetResult(),
	}
	return respBuild, nil

}
