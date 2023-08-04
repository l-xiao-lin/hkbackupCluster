package service

import (
	"errors"
	alidns20150109 "github.com/alibabacloud-go/alidns-20150109/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"hkbackupCluster/pkg/aliyundnsclient"
)

func AddDomainRecordList(AccessKeyId, AccessKeySecret, Value string, Records []string) (err error) {
	if len(Records) > 0 {
		for _, record := range Records {
			_, err = AddDomainRecord(AccessKeyId, AccessKeySecret, record, Value)
			if err != nil {
				return err
			}
		}
	} else {
		err = errors.New("records信息有误")
	}
	return
}

func AddDomainRecord(AccessKeyId, AccessKeySecret, Record, Value string) (ResponseData *alidns20150109.AddDomainRecordResponse, _err error) {
	client, _err := aliyundnsclient.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	addDomainRecordRequest := &alidns20150109.AddDomainRecordRequest{
		DomainName: tea.String("tongtool.com"),
		RR:         tea.String(Record),
		Type:       tea.String("A"),
		Value:      tea.String(Value),
		TTL:        tea.Int64(60),
	}
	runtime := &util.RuntimeOptions{}
	ResponseData, tryErr := func() (data *alidns20150109.AddDomainRecordResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		data, _err = client.AddDomainRecordWithOptions(addDomainRecordRequest, runtime)
		if _err != nil {
			return nil, _err
		}

		return ResponseData, nil
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
