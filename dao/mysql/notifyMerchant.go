package mysql

import "hkbackupCluster/logger"

type RespMerchant struct {
	TaskID int    `json:"task_id" db:"task_id"`
	Host   string `json:"host" db:"host"`
}

func GetMerchantsForIMSent() (data []RespMerchant, err error) {
	//sqlStr := "select task_id,host from release_operations " +
	//	"where im_sent=? and " +
	//	"CURRENT_TIMESTAMP >=  DATE_SUB(scheduled_time,INTERVAL 30 MINUTE)  and " +
	//	"CURRENT_TIMESTAMP < scheduled_time and " +
	//	"TIME(scheduled_time) BETWEEN '01:00:00' and '10:00:00'"

	sqlStr := "select task_id,host from release_operations " +
		"where im_sent=? and " +
		"CURRENT_TIMESTAMP >=  DATE_SUB(scheduled_time,INTERVAL 30 MINUTE)  and " +
		"TIME(scheduled_time) BETWEEN '01:00:00' and '10:00:00'"

	err = db.Select(&data, sqlStr, 0)
	if err != nil {
		logger.SugarLog.Errorf("query failed,err:%v", err)
		return
	}
	return
}
