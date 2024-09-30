package service

import (
	"encoding/json"
	"fmt"
	"strings"

	cs20151215 "github.com/alibabacloud-go/cs-20151215/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

func CreateClient(AccessKeyId, AccessKeySecret string) (_result *cs20151215.Client, _err error) {
	config := &openapi.Config{
		AccessKeyId: &AccessKeyId,

		AccessKeySecret: &AccessKeySecret,
	}
	// Endpoint 请参考 https://api.aliyun.com/product/CS
	config.Endpoint = tea.String("cs.cn-hongkong.aliyuncs.com")
	_result = &cs20151215.Client{}
	_result, _err = cs20151215.NewClient(config)
	return _result, _err
}

func CreateAskCluster(AccessKeyId, AccessKeySecret string) (respData *cs20151215.CreateClusterResponse, _err error) {
	client, _err := CreateClient(AccessKeyId, AccessKeySecret)
	if _err != nil {
		return nil, _err
	}

	addon0 := &cs20151215.Addon{
		Name: tea.String("Flannel "),
	}
	addon1 := &cs20151215.Addon{
		Name:   tea.String("nginx-ingress-controller"),
		Config: tea.String("{\"IngressSlbNetworkType\":\"internet\"}"),
	}

	createClusterRequest := &cs20151215.CreateClusterRequest{
		Name:                  tea.String("hk-ASKtongtool"),
		RegionId:              tea.String("cn-hongkong"),
		ClusterType:           tea.String("ManagedKubernetes"),
		ClusterSpec:           tea.String("ack.pro.small"),
		KubernetesVersion:     tea.String("1.30.1-aliyun.1"),
		Vpcid:                 tea.String("vpc-j6csek2yl05q1azrxamc2"),
		ContainerCidr:         tea.String("172.20.0.0/16"),
		ServiceCidr:           tea.String("10.0.0.0/16"),
		SnatEntry:             tea.Bool(true),
		EndpointPublicAccess:  tea.Bool(true),
		Timezone:              tea.String("Asia/Shanghai"),
		Addons:                []*cs20151215.Addon{addon0, addon1},
		KeyPair:               tea.String("root_bastion2"),
		VswitchIds:            []*string{tea.String("vsw-j6cvi1wl5m1n5j7w2tcqx")},
		Profile:               tea.String("Serverless"),
		ServiceDiscoveryTypes: []*string{tea.String("CoreDNS")},
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	respData, tryErr := func() (_result *cs20151215.CreateClusterResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		_result, _err = client.CreateClusterWithOptions(createClusterRequest, headers, runtime)
		if _err != nil {
			return nil, _err
		}

		return _result, nil
	}()

	if tryErr != nil {
		var error = &tea.SDKError{}
		if _t, ok := tryErr.(*tea.SDKError); ok {
			error = _t
		} else {
			error.Message = tea.String(tryErr.Error())
		}
		// 此处仅做打印展示，请谨慎对待异常处理，在工程项目中切勿直接忽略异常。
		// 错误 message
		fmt.Println(tea.StringValue(error.Message))
		// 诊断地址
		var data interface{}
		d := json.NewDecoder(strings.NewReader(tea.StringValue(error.Data)))
		d.Decode(&data)
		if m, ok := data.(map[string]interface{}); ok {
			recommend, _ := m["Recommend"]
			fmt.Println(recommend)
		}
		_, _err = util.AssertAsString(error.Message)
		if _err != nil {
			return nil, _err
		}
	}
	return respData, _err
}
