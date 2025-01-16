package pack

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/bndr/gojenkins"
)

var (
	Url            = "http://10.0.0.130:8080/jenkins/"
	deployUsername = "lixiaolin"
	deployPassword = "Sbihss46589"
)

type RespBuild struct {
	BuildNumber int64  `json:"build_number"`
	BuildResult string `json:"build_result"`
}

var ErrorConnFailed = errors.New("jenkins连接失败")

func JenkinsBuild(jobName string, param map[string]string) (respBuild *RespBuild, err error) {

	ctx := context.Background()
	jenkins := gojenkins.CreateJenkins(nil, Url, deployUsername, deployPassword)
	_, err = jenkins.Init(ctx)
	if err != nil {
		return nil, ErrorConnFailed
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

func StructToMap(p interface{}) (map[string]string, error) {
	m := make(map[string]string)
	v := reflect.ValueOf(p)
	//确保p是一个指针并且指向一个结构体
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("input must be a pointer to struct")
	}
	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)
		jsonTag := structField.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = structField.Name
		}
		var fieldValue string

		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				fieldValue = ""
			} else {
				//解引用指针并继续处理实际的值
				field := field.Elem()
				switch field.Kind() {
				case reflect.String:
					fieldValue = field.String()
				case reflect.Bool:
					fieldValue = strconv.FormatBool(field.Bool())
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					fieldValue = strconv.FormatInt(field.Int(), 10)
				case reflect.Float64, reflect.Float32:
					fieldValue = strconv.FormatFloat(field.Float(), 'f', -1, 64)
				}
			}
		} else {
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
		}
		m[jsonTag] = fieldValue
	}
	return m, nil
}
