package routers

import (
	"hkbackupCluster/controller"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
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

	return r
}
