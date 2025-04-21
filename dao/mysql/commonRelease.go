package mysql

import (
	"hkbackupCluster/logger"
	"time"
)

type respRelease struct {
	TaskID           string `json:"task_id" db:"task_id"`
	ServiceName      string `json:"service_name" db:"service_name"`
	HasConfiguration bool   `json:"has_configuration" db:"has_configuration"`
	OpenSchema       bool   `json:"open_schema" db:"open_schema"`
	Status           int    `json:"status" db:"status"`
	ScheduledTime    string `json:"scheduled_time" db:"scheduled_time"`
}

func GetCommonUnreleasedRecords() (data []respRelease, err error) {

	nowUTC := time.Now().UTC()
	sqlStr := "select task_id,open_schema,status,scheduled_time,service_name,has_configuration from common_release_operations where " +
		"status=? and scheduled_time <= ? "

	err = db.Select(&data, sqlStr, 0, nowUTC)
	if err != nil {
		logger.SugarLog.Errorf("query failed,err:%v", err)
		return
	}

	return
}

func UpdateCommonReleaseStatus(taskID string, status int, buildNumber *int64) (err error) {
	var args []interface{}
	var sqlStr string
	if buildNumber != nil {
		sqlStr = "update common_release_operations set status=? , build_number=? where task_id=?"
		args = []interface{}{status, buildNumber, taskID}
	} else {
		sqlStr = "update  common_release_operations set status=? where task_id=? "
		args = []interface{}{status, taskID}
	}
	ret, err := db.Exec(sqlStr, args...)
	if err != nil {
		logger.SugarLog.Errorf("update failed,err:%v", err)
		return
	}
	n, err := ret.RowsAffected()
	if err != nil {
		logger.SugarLog.Errorf("get RowsAffected failed,err:%v", err)
		return
	}
	if n == 0 {
		logger.SugarLog.Errorf("no rows were update for taskID:%v", taskID)
		return
	}
	return

}
