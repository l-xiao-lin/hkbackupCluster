package mysql

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
)

func InsertXml(p *model.ParamReleaseXML, status int) (err error) {
	sqlStr := "insert into release_xml (task_id,job_name,src_path,host,common,status) values (?,?,?,?,?,?)"

	ret, err := db.Exec(sqlStr, p.TaskID, p.JobName, p.SrcPath, p.Host, p.Common, status)
	if err != nil {
		logger.SugarLog.Errorf("insert task_id %s failed,err:%v", p.TaskID, err)
		return
	}
	insertID, err := ret.LastInsertId()
	if err != nil {
		logger.SugarLog.Errorf("get lastinsert ID failed,err:%v", err)
		return
	}
	logger.SugarLog.Infof("insertID:%d", insertID)
	return
}
