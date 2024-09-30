package controller

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"reflect"
	"strings"

	"hkbackupCluster/service"

	"github.com/gin-gonic/gin"
)

func structToMap(v interface{}) map[string]string {
	result := make(map[string]string)
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	t := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			key := strings.Split(jsonTag, ",")[0]
			result[key] = val.Field(i).String()
		}
	}
	return result
}

func JenkinsEasyBuildHandler(c *gin.Context) {
	p := model.NewParamsEasyseller()
	JenkinsBuildHandler(c, p)

}

func JenkinsPackageHandler(c *gin.Context) {
	p := model.NewParamsPackage()
	JenkinsBuildHandler(c, p)
}

func JenkinsBuildHandler(c *gin.Context, p model.Builder) {
	if err := p.Build(c); err != nil {
		logger.SugarLog.Errorf("invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "invalid param",
			"code": 1001,
		})
		return
	}
	paramsMap := structToMap(p)
	logger.SugarLog.Infof("paramsMap: %v", paramsMap)

	data, err := service.JenkinsBuild(paramsMap)
	if err != nil {
		logger.SugarLog.Errorf("service JenkinsBuild failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "服务繁忙",
			"code": 1002,
		})
		return
	}

	if data.Result == "FAILURE" {
		c.JSON(200, gin.H{
			"msg":  "build failed",
			"code": 1003,
			"data": data,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "success",
		"code": 1000,
		"data": data,
	})
}
