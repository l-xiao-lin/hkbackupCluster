package routers

import (
	"hkbackupCluster/controller"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/cronjobs"
	"time"

	ginzap "github.com/gin-contrib/zap"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.New()
	r.LoadHTMLGlob("templates/*")
	//使用第三方ginzap来接收gin框架的系统日志
	r.Use(ginzap.Ginzap(logger.LG, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger.LG, true))

	v1 := r.Group("/api/v1")

	v1.POST("/createAskCluster", controller.CreateAskClusterHandler)

	v1.GET("/getTaskStatus", controller.GetTaskStatusHandler)

	v1.GET("/getClusterConf", controller.GetClusterConfHandler)

	v1.GET("/getClusterSlb", controller.GetClusterSlbHandler)

	v1.POST("/addDomainRecord", controller.AddDomainRecordHandler)

	v1.POST("/getIngressIP", controller.GetIngressSlbHandler)

	v1.POST("/workflow", controller.WorkFlowHandler)

	v1.GET("/execCommand", controller.ExecCommandHandler)

	v1.POST("/uploadFile", controller.UploadFileHandler)

	v1.POST("/compareJar", controller.CompareJarHandler)

	v1.POST("/sendWeChat", controller.SendWeChatHandler)

	v1.GET("/checkSystem", controller.CheckSystemHandler)

	v1.GET("/checkListing", controller.CheckListingHandler)

	v1.GET("/getConfig", controller.GetConfigHandler)

	v1.POST("/publish", controller.PublishConfigHandler)

	v1.GET("/checkAdminYms", controller.CheckAdminYmsHandler)

	v1.GET("/checkWwwYms", controller.CheckWwwYmsHandler)

	v1.GET("/checkSupplierYms", controller.CheckSupplierYmsHandler)

	v1.POST("/jenkinsPackage", controller.JenkinsPackageHandler)

	v1.POST("/jenkinseasyseller", controller.JenkinsEasyBuildHandler)

	v1.POST("/errorCount", controller.ErpErrorCountHandler)

	v1.GET("/erpRestart", controller.RestarAndChecktHandler)

	v1.GET("/task/:task_id", controller.CheckTaskStatusHandler)

	v1.POST("/incrementalPack", controller.IncrementalPackHandler)

	v1.POST("/automateDeploy", controller.AutomateDeploymentHandler)

	v1.POST("/update-status", controller.UpdateStatusHandler)

	v1.GET("/getEnvName", controller.GetEnvNameHandler)

	v1.GET("/getInventory", controller.GetInventoryHandler)

	v1.GET("/demoPackage", cronjobs.DemoPackageHandler)

	v1.GET("/demoRelease", cronjobs.DemoReleaseHandler)

	v1.GET("/sendImMessage", cronjobs.DemoNotifyMerchantHandler)

	v1.POST("/testEnvPackage", controller.TestEnvPackageHandler)

	v1.POST("/releaseXML", controller.ReleaseXmlHandler)

	v1.GET("/domoReleaseXML", cronjobs.DemoReleaseXmlHandler)

	v1.GET("/report", controller.ReportHandler)

	v1.POST("/sendReport", controller.SendReportHandler)

	v1.POST("/sendRobotAlarm", controller.SendRobotAlarmHandler)

	v1.POST("/commonDeploy", controller.CommonDeploymentHandler)

	v1.POST("/demoPackageCommon", cronjobs.DemoCommonPackageHandler)

	v1.POST("/demoReleaseCommon", cronjobs.DemoCommonReleaseHandler)

	return r
}
