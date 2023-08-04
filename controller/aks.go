package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"hkbackupCluster/logger"
	"hkbackupCluster/service"
)

//创建ASK集群

func CreateAskClusterHandler(c *gin.Context) {
	//1、获取参数access_key_id 及获取参数access_key_id_secret
	p := new(ParamsAccessKey)
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("CreateAskClusterHandler invalid param", zap.Error(err))
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}

	//2、调用业务层
	data, err := service.CreateAskCluster(p.AccessKeyId, p.AccessKeySecret)
	if err != nil {
		panic(err)
	}

	//3、返回状态码

	c.JSON(200, gin.H{
		"msg":        "ASKCluster集群创建中...",
		"cluster_id": data.Body.ClusterId,
		"code":       1000,
		"task_id":    data.Body.TaskId,
	})

}

//获取ASK集群创建状态

func GetTaskStatusHandler(c *gin.Context) {
	//1、获取get请求参数
	p := new(ParamsTask)

	if err := c.ShouldBind(p); err != nil {
		logger.SugarLog.Errorf("GetTaskStatusHandler param invalid,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1001,
		})
		return
	}

	//2、请接业务处理
	data, err := service.GetTaskStatusInfo(p.AccessKeyId, p.AccessKeySecret, p.TaskId)
	if err != nil {
		logger.SugarLog.Errorf("service.GetTaskStatusInfo failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  fmt.Sprintf("任务处理失败,err:%v\n", err),
			"code": 1002,
		})
		return
	}

	//3、返回状态码
	c.JSON(200, gin.H{
		"code":       1000,
		"state":      data.Body.State,
		"cluster_id": data.Body.ClusterId,
		"created":    data.Body.Created,
	})

}

//获取ASK集群kube配置

func GetClusterConfHandler(c *gin.Context) {
	p := new(ParamsKubeConf)
	if err := c.ShouldBind(p); err != nil {
		logger.SugarLog.Errorf("GetClusterConfHandler param invalid,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1002,
		})
		return
	}

	if err := service.GetClusterKubeConf(p.AccessKeyId, p.AccessKeySecret, p.ClusterId); err != nil {
		logger.SugarLog.Errorf("service.GetClusterKubeConf failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1003,
		})
		return
	}
	c.JSON(200, gin.H{
		"msg":  "kubeConfig文件获取成功",
		"code": 1000,
	})

}

//获取集群slb公网IP地址

func GetClusterSlbHandler(c *gin.Context) {
	//1、获取请求参数
	p := new(ParamsSlbId)
	if err := c.ShouldBind(p); err != nil {
		logger.SugarLog.Errorf("GetClusterSlbHandler param invalid,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  "参数绑定失败",
			"code": 1002,
		})
		return
	}

	//2、调用业务层
	data, err := service.GetClusterSlbPublicIp(p.AccessKeyId, p.AccessKeySecret, p.ClusterId)
	if err != nil {
		logger.SugarLog.Errorf("service.GetClusterSlbPublicIp failed,err:%v\n", err)
		c.JSON(200, gin.H{
			"msg":  err.Error(),
			"code": 1003,
		})
		return
	}

	//3、返回状态码
	c.JSON(200, gin.H{
		"msg":  "成功获取ASK集群Slb公网IP地址",
		"code": 1000,
		"data": data,
	})
}
