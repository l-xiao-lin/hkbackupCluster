package controller

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
	"reflect"
	"strconv"

	"github.com/gin-gonic/gin"
)

func convertToMap(p interface{}) (map[string]string, error) {
	m := make(map[string]string)
	v := reflect.ValueOf(p).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		jsonTag := structField.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = structField.Name
		}
		fieldValue := ""
		switch field.Kind() {
		case reflect.String:
			fieldValue = field.String()
		case reflect.Bool:
			fieldValue = strconv.FormatBool(field.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldValue = strconv.FormatInt(field.Int(), 10)
		default:
			return nil, fmt.Errorf("unsupported type:%s", field.Type().String())
		}
		m[jsonTag] = fieldValue
	}
	return m, nil
}

func IncrementalPackHandler(c *gin.Context) {

	//1、初始化默认参数
	p := &ParamsIncrementalPack{
		Host:            "",
		Common:          "",
		Diff:            "",
		UpdateJbossConf: false,
		UpdateSdkConf:   false,
		UpdateSecurity:  false,
	}
	//2、参数校验

	if err := c.ShouldBindJSON(p); err != nil {
		logger.SugarLog.Errorf("IncrementalPackHandler invalid param,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "invalid  param",
			"code": 1001,
		})
		return
	}

	//3、将结构体的参数转换成map
	paramMap, err := convertToMap(p)
	logger.SugarLog.Infof("paramMap:%v", paramMap)
	if err != nil {
		logger.SugarLog.Errorf("convertToMap failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "convertToMap failed",
			"code": 1002,
		})
		return
	}
	//3、调用业务处理
	data, err := service.IncrementalPack(paramMap)
	if err != nil || data.BuildResult == "FAILURE" {
		logger.SugarLog.Errorf("service IncrementalPack failed,err:%v", err)
		c.JSON(200, gin.H{
			"msg":  "service IncrementalPack failed",
			"code": 1003,
		})
		return
	}

	c.JSON(200, gin.H{
		"msg":  "build success",
		"code": 1000,
		"data": data,
	})
}
