package service

import (
	"bufio"
	"errors"
	"fmt"
	cs20151215 "github.com/alibabacloud-go/cs-20151215/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/aliyunclient"
	"os"
)

var FileName string = "conf/config"

func bufferWrite(param string) (err error) {
	fileHandle, err := os.OpenFile(FileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		fmt.Printf("open file error:%v\n", err)
		return
	}
	defer fileHandle.Close()

	buf := bufio.NewWriter(fileHandle)
	buf.WriteString(param)

	err = buf.Flush()
	return

}

func GetClusterKubeConf(AccessKeyId, AccessKeySecret, ClusterId string) (err error) {
	//1、获取kube config文件
	ResponseData, err := GetClusterConf(AccessKeyId, AccessKeySecret, ClusterId)
	if err != nil {
		logger.SugarLog.Errorf("GetClusterConf failed,err:%v\n", err)
		return
	}

	//2、将文件写到本地
	if ResponseData == nil {
		return errors.New("kubeConfig文件为空")
	}

	err = bufferWrite(*ResponseData.Body.Config)
	return

}

func GetClusterConf(AccessKeyId, AccessKeySecret, ClusterId string) (ResponseData *cs20151215.DescribeClusterUserKubeconfigResponse, _err error) {
	// 请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID 和 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例使用环境变量获取 AccessKey 的方式进行调用，仅供参考，建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html
	client, _err := aliyunclient.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	describeClusterUserKubeconfigRequest := &cs20151215.DescribeClusterUserKubeconfigRequest{}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	ResponseData, tryErr := func() (data *cs20151215.DescribeClusterUserKubeconfigResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		data, _err = client.DescribeClusterUserKubeconfigWithOptions(tea.String(ClusterId), describeClusterUserKubeconfigRequest, headers, runtime)
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
	return ResponseData, nil
}
