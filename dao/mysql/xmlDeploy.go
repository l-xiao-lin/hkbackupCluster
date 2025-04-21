package mysql

import (
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"time"
)

func InsertXml(p *model.ParamReleaseXML) (err error) {

	var utcScheduledTime time.Time
	if p.ScheduledTime != "" {
		scheduledTime, _ := time.Parse(timeLayout, p.ScheduledTime)
		utcScheduledTime = scheduledTime.UTC()
	}
	sqlStr := "insert into release_xml (task_id,job_name,src_path,host,common,scheduled_time) values (?,?,?,?,?,?)" +
		"ON DUPLICATE KEY UPDATE " +
		"update_time=now()" +
		",status=values(status)"

	if p.Common != nil {
		sqlStr = fmt.Sprintf(sqlStr + ",common=values(common)")
	}

	if p.Host != "" {
		sqlStr = fmt.Sprintf(sqlStr + ",host=values(host)")
	}

	ret, err := db.Exec(sqlStr, p.TaskID, p.JobName, p.SrcPath, p.Host, p.Common, utcScheduledTime)
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

func UpdateXmlStatus(status int, taskID string) (err error) {
	sqlStr := "update release_xml set status=? where task_id=?"
	ret, err := db.Exec(sqlStr, status, taskID)
	if err != nil {
		logger.SugarLog.Errorf("UpdateXmlStatus exec failed,taskID:%s,err:%v", taskID, err)
		return err
	}
	n, err := ret.RowsAffected()
	if err != nil {
		logger.SugarLog.Errorf("UpdateXmlStatus get RowsAffected  failed,taskID:%s,err:%v", taskID, err)
		return
	}
	logger.SugarLog.Infof("UpdateXmlStatus success,affected rows:%d,taskID:%S", n, taskID)
	return nil

}

func GetUnreleasedXmlRecords() (data []model.ParamReleaseXML, err error) {
	nowUTC := time.Now().UTC()
	sqlStr := "select task_id,src_path,job_name,host,common,scheduled_time from release_xml where  status = ? and  scheduled_time <= ? "
	err = db.Select(&data, sqlStr, 0, nowUTC)
	if err != nil {
		logger.SugarLog.Errorf("GetUnreleasedXmlRecords query failed,err:%v", err)
		return
	}
	return
}
