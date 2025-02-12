package service

import (
	"hkbackupCluster/dao/mysql"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
)

func ReleaseXml(p *model.ParamReleaseXML) (err error) {
	if err := mysql.InsertXml(p); err != nil {
		logger.SugarLog.Errorf("mysql.InsertXml failed,task_id:%s,err:%v", p.TaskID, err)
		return err
	}
	return

}
