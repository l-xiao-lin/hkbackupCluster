package service

import (
	cs20151215 "github.com/alibabacloud-go/cs-20151215/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/aliyunclient"
	"time"
)

func GetTaskStatusInfo(AccessKeyId, AccessKeySecret, taskId string) (ResponseData *cs20151215.DescribeTaskInfoResponse, _err error) {

	for i := 0; i < 30; i++ {
		time.Sleep(time.Second * 20)
		logger.SugarLog.Infof("检测状态第%d次", i+1)
		ResponseData, _err = GetTaskStatus(AccessKeyId, AccessKeySecret, taskId)

		if _err != nil {
			return nil, _err
		}

		if ResponseData == nil {
			logger.SugarLog.Errorf("ResponseData is nil")
			continue
		}

		if *ResponseData.Body.State == "success" {
			break
		}

	}
	return ResponseData, nil

}

func GetTaskStatus(AccessKeyId, AccessKeySecret, taskId string) (ResponseData *cs20151215.DescribeTaskInfoResponse, _err error) {

	client, _err := aliyunclient.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	ResponseData, tryErr := func() (data *cs20151215.DescribeTaskInfoResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		data, _err = client.DescribeTaskInfoWithOptions(tea.String(taskId), headers, runtime)
		if _err != nil {
			return nil, _err
		}

		return data, nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 如有需要，请打印 error
		_, _err = util.AssertAsString(error.Message)
		if _err != nil {
			return nil, _err
		}
	}
	return ResponseData, _err
}
