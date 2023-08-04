package main

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/routers"
)

func main() {
	if err := logger.Init("dev"); err != nil {
		fmt.Printf("init logger failed,err:%v\n", err)
		return
	}

	r := routers.SetupRouter()

	r.Run(":8000")
}
