package service

import (
	"context"
	"hkbackupCluster/logger"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const kubeconfigPath = "./conf/config"

var (
	namespace   = "kube-system"
	serviceName = "nginx-ingress-lb"
)

func GetIngressPublicIP() (*string, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		logger.SugarLog.Errorf("clientcmd BuildConfigFromFlags failed,err:%v", err)
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.SugarLog.Errorf("NewForConfig failed,err:%v", err)
		return nil, err
	}

	service, err := clientSet.CoreV1().Services(namespace).Get(context.TODO(), serviceName, v1.GetOptions{})
	if err != nil {
		logger.SugarLog.Errorf("get service failed,err:%v", err)
		return nil, err
	}

	if len(service.Status.LoadBalancer.Ingress) > 0 {
		ip := service.Status.LoadBalancer.Ingress[0].IP
		if ip != "" {
			logger.SugarLog.Infof("loadBalance IP:%v", ip)
			return &ip, nil
		}
	}
	logger.SugarLog.Errorf("not found ingress slb public:%v", err)
	return nil, err

}
