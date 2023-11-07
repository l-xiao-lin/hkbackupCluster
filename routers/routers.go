package routers

import (
	"github.com/gin-gonic/gin"
	"hkbackupCluster/controller"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()
	v1 := r.Group("/api/v1")

	v1.POST("/createAskCluster", controller.CreateAskClusterHandler)

	v1.GET("/getTaskStatus", controller.GetTaskStatusHandler)

	v1.GET("/getClusterConf", controller.GetClusterConfHandler)

	v1.GET("/getClusterSlb", controller.GetClusterSlbHandler)

	v1.POST("/addDomainRecord", controller.AddDomainRecordHandler)

	v1.POST("/workflow", controller.WorkFlowHandler)

	v1.GET("/execCommand", controller.ExecCommandHandler)

	v1.POST("/uploadFile", controller.UploadFileHandler)

	v1.POST("/compareJar", controller.CompareJarHandler)

	return r
}
