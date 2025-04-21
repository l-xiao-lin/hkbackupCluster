package service

import (
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/model"
)

func CommonDeployment(p *model.ParamCommonDeploy) (err error) {
	return mysql.InsertCommonService(p)

}
