package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"time"
)

func InsertCommonService(param *model.ParamCommonDeploy) (err error) {
	tx, err := db.Begin()
	if err != nil {
		logger.SugarLog.Errorf("InsertCommon begin trans failed,err:%v", err)
		return
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			logger.SugarLog.Errorf("InsertCommonService  task_id:%s rollback", param.TaskID)
			tx.Rollback()
		} else {
			tx.Commit()
			logger.SugarLog.Infof("InsertCommon commit")

		}
	}()

	//将接收到的参数写进common_package_operations表中
	if err := insertRecord(tx, param); err != nil {
		return err
	}

	//如果存在配置文件,则写进common_package_configurations表中
	if param.HasConfiguration && len(param.CommonConfigs) > 0 {
		if err := insertCommonConfigurations(tx, param); err != nil {
			return err
		}
	}
	return

}

func insertRecord(tx *sql.Tx, p *model.ParamCommonDeploy) (err error) {
	sqlStr := "insert into common_package_operations(task_id,service_name,has_configuration,status,package_time,scheduled_time,open_schema)" +
		"values(?,?,?,?,?,?,?)" +
		"ON DUPLICATE KEY UPDATE " +
		"update_time=now()" +
		",status=values(status)"

	has_configuration := 0
	if p.HasConfiguration {
		has_configuration = 1
		sqlStr = fmt.Sprintf(sqlStr + ",has_configuration=values(has_configuration)")
	}

	openSchema := 0
	if p.OpenSchema {
		openSchema = 1
		sqlStr = fmt.Sprintf(sqlStr + ",open_schema=values(open_schema)")
	}

	var utcPackageTime time.Time
	if p.PackageTime != "" {
		packageTime, _ := time.Parse(timeLayout, p.PackageTime)
		utcPackageTime = packageTime.UTC()
		sqlStr = fmt.Sprintf(sqlStr + ",package_time=values(package_time)")
	}

	if p.ServiceName != "" {
		sqlStr = fmt.Sprintf(sqlStr + ",service_name=values(service_name)")
	}

	var utcScheduledTime time.Time
	if p.ScheduledTime != "" {
		scheduledTime, _ := time.Parse(timeLayout, p.ScheduledTime)
		utcScheduledTime = scheduledTime.UTC()
		sqlStr = fmt.Sprintf(sqlStr + ",scheduled_time=values(scheduled_time)")
	}
	logger.SugarLog.Infof("sqlstr: %s", sqlStr)
	ret, err := tx.Exec(sqlStr, p.TaskID, p.ServiceName, has_configuration, p.Status, utcPackageTime, utcScheduledTime, openSchema)
	if err != nil {
		logger.SugarLog.Errorf("insert failed,err:%v", err)
		return
	}
	insertID, err := ret.LastInsertId()
	if err != nil {
		logger.SugarLog.Errorf("get lastinsert ID failed,err:%v", err)
		return
	}
	logger.SugarLog.Infof("insert success,insertID: %d", insertID)
	return

}

func insertCommonConfigurations(tx *sql.Tx, p *model.ParamCommonDeploy) (err error) {
	sqlStr := "insert into common_package_configurations(task_id,service_name,config_action,config_content)" +
		"values(?,?,?,?)"

	for _, config := range p.CommonConfigs {
		_, err = tx.Exec(sqlStr, p.TaskID, p.ServiceName, config.Action, config.Content)
		if err != nil {
			logger.SugarLog.Errorf("insertCommonConfigurations taskID:%s insert failed,err:%v", p.TaskID, err)
			return
		}
	}
	return

}

func GetUnPackageCommonRecords() (data []model.RespPackageCommonData, err error) {
	nowUTC := time.Now().UTC()
	sqlStr := "select task_id,service_name,has_configuration,status,package_time,scheduled_time,open_schema from common_package_operations " +
		"where status=? and package_time <=? order by create_time ASC "
	err = db.Select(&data, sqlStr, 0, nowUTC)
	if err != nil {
		logger.SugarLog.Errorf("GetUnPackageCommonRecords failed,err:%v", err)
		return
	}
	return
}

func UpdatePackCommonStatus(taskID string, status int, buildNumber *int64) (err error) {
	var args []interface{}
	var sqlStr string
	if buildNumber != nil {
		sqlStr = "update common_package_operations  set status=?,build_number=? where task_id=? "
		args = []interface{}{status, *buildNumber, taskID}
	} else {
		sqlStr = "update common_package_operations set status=? where task_id=?"
		args = []interface{}{status, taskID}
	}
	ret, err := db.Exec(sqlStr, args...)
	if err != nil {
		logger.SugarLog.Errorf("UpdatePackCommonStatus taskID:%s failed,err:%v", taskID, err)
		return
	}
	n, err := ret.RowsAffected()
	if err != nil {
		logger.SugarLog.Errorf("get RowsAffected failed,err:%v", err)
		return
	}
	logger.SugarLog.Infof("RowsAffected %d record", n)
	return
}

func GetCommonConfigurations(taskID string) (data []model.RespConfCommon, err error) {
	sqlStr := "select task_id,service_name,config_action,config_content from common_package_configurations where task_id=? order by create_time ASC"
	err = db.Select(&data, sqlStr, taskID)
	if err != nil {
		logger.SugarLog.Errorf("GetCommonConfigurations failed,err:%v", err)
		return
	}
	return
}

func HandleSuccessCommonPackTransaction(ID string, status int, buildNumber *int64, p *model.RespPackageCommonData) (err error) {
	tx, err := db.Begin()
	if err != nil {
		logger.SugarLog.Errorf("begin trans failed,err:%v", err)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			logger.SugarLog.Errorf("rollback")
			tx.Rollback()
		} else {
			err = tx.Commit()
			logger.SugarLog.Info("commit")
		}
	}()

	var buildNum int64
	if buildNumber != nil {
		buildNum = *buildNumber
	} else {
		buildNum = 0
	}
	//更新common_package_operations表状态
	sqlStr1 := "update common_package_operations set status=?,build_number=? where task_id=? "
	ret1, err := tx.Exec(sqlStr1, status, buildNum, ID)
	if err != nil {
		return
	}
	n, err := ret1.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed")
	}

	//写入common_release_operations表记录
	scheduledTime, _ := time.Parse(time.RFC3339, p.ScheduledTime)
	utcScheduledTime := scheduledTime.UTC()

	openSchema := 0
	if p.OpenSchema {
		openSchema = 1
	}

	sqlStr2 := "insert into common_release_operations(task_id,scheduled_time,open_schema,service_name,has_configuration)values(?,?,?,?,?)"
	ret2, err := tx.Exec(sqlStr2, p.TaskID, utcScheduledTime, openSchema, p.ServiceName, p.HasConfiguration)
	if err != nil {
		return
	}
	insertID, err := ret2.LastInsertId()
	if err != nil {
		return
	}
	logger.SugarLog.Infof("insert into common_release_operations row success,id:%d", insertID)
	return

}
