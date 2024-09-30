package service

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"hkbackupCluster/logger"
	"os"
)

var configClient config_client.IConfigClient

func InitClient(namespace string) (err error) {
	//创建客户端配置
	clientConfig := constant.ClientConfig{
		TimeoutMs:           500,
		BeatInterval:        0,
		NamespaceId:         namespace,
		UpdateThreadNum:     20,
		NotLoadCacheAtStart: true,
		Username:            "nacos",
		Password:            "Okn1Wf834yxXiIFn",
	}

	//创建服务端配置
	serverConfigs := []constant.ServerConfig{
		{Scheme: "http",
			ContextPath: "/nacos",
			IpAddr:      "nacos.tongtool.com",
			Port:        8848,
		},
	}

	//创建配置客户端

	configClient, err = clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig:  &clientConfig,
		ServerConfigs: serverConfigs,
	})

	if err != nil {
		logger.SugarLog.Errorf("Failed to create Nacos config client:%v", err)
		return
	}
	return
}

func PublishConfig(namespace, dataID, group, content string) (err error) {
	//初始化client端

	err = InitClient(namespace)
	if err != nil {
		logger.SugarLog.Errorf("InitClient failed,err:%v", err)
		return
	}

	_, err = configClient.PublishConfig(vo.ConfigParam{
		DataId:  dataID,
		Group:   group,
		Content: content,
	})
	if err != nil {
		logger.SugarLog.Errorf("publish Config failed:%v", err)
		return
	}
	return

}

func GetConfig(namespace, dataID, group string) (err error) {
	//初始化client端

	err = InitClient(namespace)
	if err != nil {
		logger.SugarLog.Errorf("InitClient failed,err:%v", err)
		return
	}

	//获取配置文件

	content, err := configClient.GetConfig(vo.ConfigParam{
		DataId: dataID,
		Group:  group,
	})

	if err != nil {
		logger.SugarLog.Errorf("Failed to get config from Nacos:%v", err)
		return
	}
	fmt.Println("Config content:", content)

	//将获取到的配置内容写到文件中

	file, err := os.OpenFile(dataID, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666)
	if err != nil {
		logger.SugarLog.Errorf("open file failed:%v", err)
		return
	}
	defer file.Close()
	file.Write([]byte(content))

	return
}
