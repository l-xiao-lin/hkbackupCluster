package service

import (
	cs20151215 "github.com/alibabacloud-go/cs-20151215/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"hkbackupCluster/pkg/aliyunclient"
)

func CreateAskCluster(AccessKeyId, AccessKeySecret string) (ResponseData *cs20151215.CreateClusterResponse, _err error) {

	client, _err := aliyunclient.CreateClient(tea.String(AccessKeyId), tea.String(AccessKeySecret))
	if _err != nil {
		return nil, _err
	}

	addon0 := &cs20151215.Addon{
		Name: tea.String("nginx-ingress-controller"),

		Config: tea.String("{\"IngressSlbNetworkType\":\"internet\",\"IngressSlbSpec\":\"slb.s2.small\"}"),
	}
	workerDataDisks0 := &cs20151215.CreateClusterRequestWorkerDataDisks{
		Category: tea.String("cloud_essd"),
		Size:     tea.String("120"),
	}
	createClusterRequest := &cs20151215.CreateClusterRequest{
		Name:                      tea.String("hk-ASKtongtool"),
		RegionId:                  tea.String("cn-hongkong"),
		ClusterType:               tea.String("ManagedKubernetes"),
		ClusterSpec:               tea.String("ack.pro.small"),
		KubernetesVersion:         tea.String("1.26.3-aliyun.1"),
		ServiceCidr:               tea.String("10.0.0.0/16"),
		IsEnterpriseSecurityGroup: tea.Bool(true),
		SnatEntry:                 tea.Bool(true),
		EndpointPublicAccess:      tea.Bool(true),
		Timezone:                  tea.String("Asia/Shanghai"),
		LoginPassword:             tea.String("otPvhfv%@u92nDEP"),
		MasterSystemDiskCategory:  tea.String("cloud_efficiency"),
		MasterInstanceTypes:       []*string{tea.String("ecs.g8a.xlarge")},
		MasterSystemDiskSize:      tea.Int64(120),
		NumOfNodes:                tea.Int64(3),
		WorkerInstanceTypes:       []*string{tea.String("ecs.g8a.xlarge")},
		WorkerSystemDiskCategory:  tea.String("cloud_efficiency"),
		WorkerSystemDiskSize:      tea.Int64(120),
		WorkerDataDisks:           []*cs20151215.CreateClusterRequestWorkerDataDisks{workerDataDisks0},
		KeyPair:                   tea.String("root_bastion2"),
		WorkerVswitchIds:          []*string{tea.String("vsw-j6cvi1wl5m1n5j7w2tcqx")},
		MasterVswitchIds:          []*string{tea.String("vsw-j6cvi1wl5m1n5j7w2tcqx")},
		VswitchIds:                []*string{tea.String("vsw-j6cvi1wl5m1n5j7w2tcqx")},
		Vpcid:                     tea.String("vpc-j6csek2yl05q1azrxamc2"),
		ContainerCidr:             tea.String("172.20.0.0/16"),
		Profile:                   tea.String("Serverless"),
		Addons:                    []*cs20151215.Addon{addon0},
		ServiceDiscoveryTypes:     []*string{tea.String("CoreDNS")},
		LoadBalancerSpec:          tea.String("slb.s2.small"),
	}
	runtime := &util.RuntimeOptions{}
	headers := make(map[string]*string)
	ResponseData, tryErr := func() (data *cs20151215.CreateClusterResponse, _e error) {
		defer func() {
			if r := tea.Recover(recover()); r != nil {
				_e = r
			}
		}()
		// 复制代码运行请自行打印 API 的返回值
		result, _err := client.CreateClusterWithOptions(createClusterRequest, headers, runtime)

		if _err != nil {
			return nil, _err
		}
		return result, nil
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
