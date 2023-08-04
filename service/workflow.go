package service

import (
	"errors"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
)

var Cmd = "kubectl --kubeconfig=/home/tomcat/.kube/config apply -f /home/tomcat/hkaskCluster/HKinitResource.yml"

func WorkFlow(p *model.WorkFlow) (err error) {
	//1、创建集群,获取集群ID 集群任务

	ResponseCluster, err := CreateAskCluster(p.AccessKeyId, p.AccessKeySecret)
	if err != nil {
		logger.SugarLog.Errorf("CreateAskCluster failed,err:%v\n", err)
		return
	}
	if ResponseCluster == nil {
		err = errors.New("ASK集群创建失败")
		return
	}
	logger.SugarLog.Infof("ASK集群初始化中...")

	clusterID := ResponseCluster.Body.ClusterId
	taskID := ResponseCluster.Body.TaskId

	logger.SugarLog.Infof("clusterID:%s\n", *clusterID)
	logger.SugarLog.Infof("taskID:%s\n", *taskID)

	//2、监听集群是否创建完成

	_, err = GetTaskStatusInfo(p.AccessKeyId, p.AccessKeySecret, *taskID)
	if err != nil {
		logger.SugarLog.Errorf("GetTaskStatusInfo failed,err:%v\n", err)
		return
	}

	logger.SugarLog.Infof("集群初始化完成")

	//3、获取kubeConf配置文件

	if err = GetClusterKubeConf(p.AccessKeyId, p.AccessKeySecret, *clusterID); err != nil {
		return
	}

	logger.SugarLog.Infof("获取kubeConfig文件完成")

	//4、获取集群负载均衡公网IP地址

	PublicIP, err := GetClusterSlbPublicIp(p.AccessKeyId, p.AccessKeySecret, *clusterID)

	if err != nil {
		return
	}
	logger.SugarLog.Infof("获取集群SLB公网ip完成")

	//5、添加A记录

	err = AddDomainRecordList(p.AccessKeyId, p.AccessKeySecret, *PublicIP, p.Records)
	if err != nil {
		logger.SugarLog.Errorf("AddDomainRecord failed,err:%v\n", err)
		return
	}

	logger.SugarLog.Infof("修改A记录成功")

	//6、上传kubeConfig文件至远程机器

	if err = UploadFile(p.LocalFilePath, p.RemoteDir, p.RemoteHost); err != nil {
		logger.SugarLog.Errorf("UploadFile failed,err:%v\n", err)
		return
	}
	logger.SugarLog.Infof("UploadFile上传成功")

	//7、执行ASK资源
	data, err := ExecCommand(Cmd, p.RemoteHost)
	if err != nil {
		logger.SugarLog.Errorf("ExecCommand failed,err:%v\n", err)
		return
	}
	logger.SugarLog.Infof("集群yaml资源初始化成功,%v\n", data)

	return
}
