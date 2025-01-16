package service

import (
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/model"
)

func AutomateDeployment(p *model.ParamsIncrementalPack) (insertID int64, err error) {
	//1、将接收到的参数写进表中
	insertID, err = mysql.AutomateDeployment(p)
	if err != nil {
		return
	}

	//2、返回状态
	return insertID, nil

}

func UpdateStatus(p *model.ParamsUpdateStatus) (err error) {
	return mysql.UpdateReleaseStatus(p.TaskID, p.Status, nil)
}
