package service

import (
	"errors"
	cs20151215 "github.com/alibabacloud-go/cs-20151215/v3/client"
	slb20140515 "github.com/alibabacloud-go/slb-20140515/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/aliyunclient"
	"hkbackupCluster/pkg/aliyunclientv4"
)

func GetClusterSlbPublicIp(AccessKeyId, AccessKeySecret, ClusterId string) (PublicIp *string, err error) {

	//1、调用ASK集群状态接口,生成负载均衡id
	data, err := GetClusterSlb(AccessKeyId, AccessKeySecret, ClusterId)
	if err != nil {
		logger.SugarLog.Error("GetClusterSlb failed,err:%v\n", err)
		return nil, err
	}

	if data == nil {
		err = errors.New("GetClusterSlb 返回的数据为空")
		return nil, err
	}

	PublicInstanceId := data.Body.ExternalLoadbalancerId
	logger.SugarLog.Infof("PublicInstanceId %s\n", *PublicInstanceId)

	//2、调用slb详情接口，生成slb公网IP地址
	LbResponseData, err := GetPublicIpByInstanceId(AccessKeyId, AccessKeySecret, *PublicInstanceId)
	if err != nil {
		logger.SugarLog.Error("GetPublicIpByInstanceId failed,err:%v\n", err)
		return nil, err
	}

	if LbResponseData == nil {
		err = errors.New("GetPublicIpByInstanceId返回的数据为空")
		return nil, err
	}
	PublicIp = LbResponseData.Body.Address

	logger.SugarLog.Infof("PublicIp:%s\n", *PublicIp)
	return

}

func GetPublicIpByInstanceId(AccessKeyId, AccessKeySecret string, InstanceId string) (ResponseData *slb20140515.DescribeLoadBalancerAttributeResponse, err error) {
	client, _err := aliyunclientv4.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	describeLoadBalancerAttributeRequest := &slb20140515.DescribeLoadBalancerAttributeRequest{
		LoadBalancerId: tea.String(InstanceId),
	}
	runtime := &util.RuntimeOptions{}
	ResponseData, tryErr := func() (data *slb20140515.DescribeLoadBalancerAttributeResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		data, _err = client.DescribeLoadBalancerAttributeWithOptions(describeLoadBalancerAttributeRequest, runtime)
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

func GetClusterSlb(AccessKeyId, AccessKeySecret, ClusterId string) (ResponseData *cs20151215.DescribeClusterDetailResponse, _err error) {
	client, _err := aliyunclient.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	ResponseData, tryErr := func() (data *cs20151215.DescribeClusterDetailResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		data, _err = client.DescribeClusterDetailWithOptions(tea.String(ClusterId), headers, runtime)
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
