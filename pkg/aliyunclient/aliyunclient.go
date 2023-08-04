package aliyunclient

import (
	cs20151215 "github.com/alibabacloud-go/cs-20151215/v3/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

func CreateClient(accessKeyId *string, accessKeySecret *string) (_result *cs20151215.Client, _err error) {
	config := &openapi.Config{
		// 必填，您的 AccessKey ID
		AccessKeyId: accessKeyId,
		// 必填，您的 AccessKey Secret
		AccessKeySecret: accessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/CS
	config.Endpoint = tea.String("cs.cn-hongkong.aliyuncs.com")
	_result = &cs20151215.Client{}
	_result, _err = cs20151215.NewClient(config)
	return _result, _err
}
