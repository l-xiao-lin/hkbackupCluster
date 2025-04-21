package cronjobs

import (
	"hkbackupCluster/logger"
	"time"

	"github.com/robfig/cron/v3"
)

type TaskRunner struct {
	cron *cron.Cron
}

func NewTaskRunner() *TaskRunner {
	return &TaskRunner{cron: cron.New()}
}

func (tr *TaskRunner) Start() {
	tr.cron.AddFunc("*/5 * * * *", func() {
		if err := RunPackageTask(); err != nil {
			logger.SugarLog.Errorf("RunPackageTask failed at %s,err:%v", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	})
	tr.cron.AddFunc("*/5 * * * * ", func() {
		if err := RunReleaseTask(); err != nil {
			logger.SugarLog.Errorf("RunReleaseTask failed at %s,err:%v", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	})

	tr.cron.AddFunc("*/5 * * * * ", func() {
		if err := ReleaseXmlHandler(); err != nil {
			logger.SugarLog.Errorf("ReleaseXmlHandler failed at %s,err;%v", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	})

	tr.cron.AddFunc("*/5 * * * * ", func() {
		if err := RunCommonPackageTask(); err != nil {
			logger.SugarLog.Errorf("RunCommonPackageTask failed at %s,err;%v", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	})

	tr.cron.AddFunc("*/5 * * * * ", func() {
		if err := RunCommonReleaseTask(); err != nil {
			logger.SugarLog.Errorf("RunCommonReleaseTask failed at %s,err;%v", time.Now().Format("2006-01-02 15:04:05"), err)
		}
	})

	//tr.cron.AddFunc("*/5 * * * * ", func() {
	//	if err := notifyMerchant(); err != nil {
	//		logger.SugarLog.Errorf("notifyMerchant failed at %s,err;%v", time.Now().Format("2006-01-02 15:04:05"), err)
	//	}
	//})

	tr.cron.Start()
}

func (tr *TaskRunner) Stop() {
	tr.cron.Stop()
}
