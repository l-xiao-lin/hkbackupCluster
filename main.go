package main

import (
	"fmt"
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/pkg/cronjobs"
	"hkbackupCluster/routers"
	"hkbackupCluster/settings"
)

func main() {

	if err := logger.Init("dev"); err != nil {
		fmt.Printf("init logger failed,err:%v\n", err)
		return
	}
	if err := settings.Init(); err != nil {
		fmt.Printf("settings init failed,err:%v\n", err)
		return
	}

	if err := mysql.Init(); err != nil {
		fmt.Printf("mysql Init failed,err:%v\n", err)
		return
	}
	defer mysql.Close()

	taskRunner := cronjobs.NewTaskRunner()
	taskRunner.Start()
	defer taskRunner.Stop()

	r := routers.SetupRouter()
	r.Run(":8000")

}
