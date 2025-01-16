package mysql

import (
	"fmt"
	"hkbackupCluster/logger"
	"time"
)

type ReleaseOperation struct {
	TaskID        string  `db:"task_id" json:"task_id"`
	Host          string  `db:"host" json:"host"`
	Status        int8    `db:"status" json:"status"`
	RmRulePackage bool    `db:"rm_rulepackage" json:"rm_rulepackage"`
	PkgName       *string `db:"pkg_name" json:"pkg_name"`
	IsSqlExec     bool    `db:"is_sql_exec" json:"is_sql_exec"`
}

func GetUnreleasedRecords() (data []ReleaseOperation, err error) {
	nowUTC := time.Now().UTC()

	logger.SugarLog.Infof("nowUTC:%v", nowUTC)

	sqlStr := "select task_id,host,rm_rulepackage,pkg_name,is_sql_exec,status from release_operations " +
		"where (status = ? or status= ?  or status= ?)and scheduled_time<= ?"
	fmt.Printf("sqlstr: %s\n", sqlStr)
	err = db.Select(&data, sqlStr, 0, 4, 3, nowUTC)

	if err != nil {
		logger.SugarLog.Errorf("query failed,err:%v", err)
		return
	}
	return
}

func UpdateReleaseStatus(taskID string, status int8, buildNumber *int64) (err error) {
	var args []interface{}
	var sqlStr string
	if buildNumber != nil {
		sqlStr = "update release_operations set status=? , build_number=? where task_id=?"
		args = []interface{}{status, buildNumber, taskID}
	} else {
		sqlStr = "update  release_operations set status=? where task_id=? "
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
