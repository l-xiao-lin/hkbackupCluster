package aliyunclientv4

import (
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	slb20140515 "github.com/alibabacloud-go/slb-20140515/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *slb20140515.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/Slb
	config.Endpoint = tea.String("slb.aliyuncs.com")
	_result = &slb20140515.Client{}
	_result, _err = slb20140515.NewClient(config)
	return _result, _err
}
