package mysql

import "hkbackupCluster/logger"

type RespMerchant struct {
	TaskID string `json:"task_id" db:"task_id"`
	Host   string `json:"host" db:"host"`
}

func GetMerchantsForIMSent() (data []RespMerchant, err error) {
	sqlStr := "select task_id,host from release_operations " +
		"where im_send=? and should_send=?  and " +
		"CURRENT_TIMESTAMP >=  DATE_SUB(scheduled_time,INTERVAL 10 MINUTE)  and " +
		"CURRENT_TIMESTAMP < scheduled_time and " +
		"TIME(scheduled_time) BETWEEN '01:00:00' and '10:00:00'"
	err = db.Select(&data, sqlStr, 0, 1)
	if err != nil {
		logger.SugarLog.Errorf("query failed,err:%v", err)
		return
	}
	return
}

func UpdateIMSentStatus(IMStatus int, taskID string) (err error) {
	sqlStr := "update release_operations set im_send=? where task_id=?"
	result, err := db.Exec(sqlStr, IMStatus, taskID)
	if err != nil {
		logger.SugarLog.Errorf("UpdateIMSentStatus exec failed,taskID:%s,err:%v", taskID, err)
		return
	}
	n, err := result.RowsAffected()
	if err != nil {
		return
	}
	logger.SugarLog.Infof("UpdateIMSentStatus success,affected rows:%d,taskID:%s", n, taskID)
	return
}
