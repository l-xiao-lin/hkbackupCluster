package main

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/routers"
	"hkbackupCluster/settings"
)

func main() {

	if err := logger.Init("dev"); err != nil {
		fmt.Printf("init logger failed,err:%v\n", err)
		return
	}

	if err := settings.Init(); err != nil {
		fmt.Printf("settings init failed,err:%v", err)
		return
	}

	r := routers.SetupRouter()

	r.Run(":8000")
}
