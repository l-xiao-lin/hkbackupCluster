package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"time"
)

const timeLayout = "2006-01-02T15:04:05-07:00"

func AutomateDeployment(p *model.ParamsIncrementalPack) (insertID int64, err error) {
	tx, err := db.Begin()
	if err != nil {
		logger.SugarLog.Errorf("AutomateDeployment begin trans failed,err:%v", err)
		return 0, err
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
			logger.SugarLog.Infof("AutomateDeployment commit")
		}
	}()

	//将参数写进package_operations表中
	insertID, err = InsertPackage(tx, p)
	if err != nil {
		return 0, err
	}

	//如果存在配置参数,则进行配置表中

	err = insertConfigurations(tx, p)

	return insertID, nil
}

func insertConfigurations(tx *sql.Tx, p *model.ParamsIncrementalPack) (err error) {
	sqlStr := "insert into package_configurations(task_id,config_type,config_content,config_action,host) values(?,?,?,?,?)"
	//在插入数据前,删除taskID对应的配置文件，以防止重复的配置文件
	sqlStr1 := "delete from package_configurations where task_id=?"
	_, err = db.Exec(sqlStr1, p.TaskID)
	if err != nil {
		return err
	}
	if p.UpdateJbossConf && len(p.JbossConf) > 0 {
		for _, config := range p.JbossConf {
			_, err = tx.Exec(sqlStr, p.TaskID, config.ConfigType, config.ConfigContent, config.ConfigAction, p.Host)
			if err != nil {
				logger.SugarLog.Errorf("failed to insert jboss configuration:%v", err)
				return
			}
		}
	}
	if p.UpdateSdkConf && len(p.SdkConf) > 0 {
		for _, config := range p.SdkConf {
			_, err = tx.Exec(sqlStr, p.TaskID, config.ConfigType, config.ConfigContent, config.ConfigAction, p.Host)
			if err != nil {
				logger.SugarLog.Errorf("failed to insert sdk configuration:%v", err)
				return
			}

		}
	}
	return

}

func InsertPackage(tx *sql.Tx, p *model.ParamsIncrementalPack) (insertID int64, err error) {
	sqlStr := "insert into package_operations(task_id,job_name,host,status,src_path,common,diff,rm_rulepackage, pkg_name,update_jbossconf,update_sdkconf,update_security,package_time,scheduled_time,is_sql_exec,is_package,canary_status) " +
		"values(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) " +
		"ON DUPLICATE KEY UPDATE " +
		"update_time=now()" +
		",status=values(status)"
	update_jbossconf := 0
	if p.UpdateJbossConf {
		update_jbossconf = 1
	}
	update_sdkconf := 0
	if p.UpdateSdkConf {
		update_sdkconf = 1
	}
	update_security := 0
	if p.UpdateSecurity {
		update_security = 1
	}
	is_sql_exec := 0
	if p.IsSqlExec {
		is_sql_exec = 1
	}
	rm_rulepackage := 0
	if p.RmRulepackage {
		rm_rulepackage = 1
	}

	is_package := 0
	if p.IsPackage {
		is_package = 1
	}

	var status int
	if p.Status == 3 {
		status = 3
	}
	var utcPackageTime time.Time
	if p.PackageTime != "" {
		//解析带有时区信息的时间字符串
		packageTime, _ := time.Parse(timeLayout, p.PackageTime)
		utcPackageTime = packageTime.UTC()
		sqlStr = fmt.Sprintf(sqlStr + ",package_time=values(package_time)")
	}

	var utcScheduledTime time.Time
	if p.ScheduledTime != "" {
		scheduledTime, _ := time.Parse(timeLayout, p.ScheduledTime)
		utcScheduledTime = scheduledTime.UTC()
		sqlStr = fmt.Sprintf(sqlStr + ",scheduled_time=values(scheduled_time)")
	}

	if p.Host != "" {
		sqlStr = fmt.Sprintf(sqlStr + ",host=values(host)")
	}
	if p.SrcPath != "" {
		sqlStr = fmt.Sprintf(sqlStr + ",src_path=values(src_path)")
	}

	if p.JobName != "" {
		sqlStr = fmt.Sprintf(sqlStr + ",job_name=values(job_name)")
	}

	if p.CanaryStatus != nil {
		sqlStr = fmt.Sprintf(sqlStr + ",canary_status=values(canary_status)")
	}

	fmt.Printf("sqlStr:%s\n", sqlStr)
	ret, err := tx.Exec(sqlStr, p.TaskID, p.JobName, p.Host, status, p.SrcPath, p.Common, p.Diff, rm_rulepackage, p.PkgName, update_jbossconf, update_sdkconf, update_security, utcPackageTime, utcScheduledTime, is_sql_exec, is_package, p.CanaryStatus)
	if err != nil {
		logger.SugarLog.Errorf("insert failed,err:%v", err)
		return
	}
	insertID, err = ret.LastInsertId()
	if err != nil {
		logger.SugarLog.Errorf("get lastinsert ID failed,err:%v", err)
		return
	}
	logger.SugarLog.Infof("insert success,the ID is %d", insertID)
	return

}

func GetUnPackageRecords() (data []model.RespPackageData, err error) {
	nowUTC := time.Now().UTC()
	sqlStr := "select task_id,job_name,host,common,diff,rm_rulepackage,src_path,pkg_name,update_jbossconf,update_sdkconf,update_security,package_time,scheduled_time,is_sql_exec,is_package,canary_status from package_operations " +
		"where status=? and package_time<=? order by create_time ASC"
	err = db.Select(&data, sqlStr, 0, nowUTC)
	if err != nil {
		logger.SugarLog.Errorf("GetUnPackageRecords failed,err:%v", err)
		return
	}
	return
}

func GetConfigurations(taskID string) (data []model.Config, err error) {
	sqlStr := "select config_type,config_content,config_action,host from package_configurations where task_id=? order by create_time ASC"
	err = db.Select(&data, sqlStr, taskID)
	if err != nil {
		logger.SugarLog.Errorf("GetConfigurations failed,err:%v", err)
		return
	}
	return

}

func UpdatePackStatus(Id string, status int8, buildNumber *int64) (err error) {
	var args []interface{}
	var sqlStr string
	if buildNumber != nil {
		sqlStr = "update package_operations set status=?,build_number=? where task_id=?"
		args = []interface{}{status, *buildNumber, Id}
	} else {
		sqlStr = "update package_operations set status=? where task_id=? "
		args = []interface{}{status, Id}
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
		return fmt.Errorf("no rows were updated for id:%d", n)
	}
	return

}

// HandleSuccessPackTransaction 打包成功后更新状态并写到release_operations表中
func HandleSuccessPackTransaction(Id string, status int8, buildNumber *int64, p *model.RespPackageData) (err error) {
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

	//更新package_operations状态
	sqlStr1 := "update package_operations set status=?,build_number=? where task_id=?"
	ret1, err := tx.Exec(sqlStr1, status, buildNum, Id)
	if err != nil {
		return
	}
	n, err := ret1.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		return errors.New("exec sqlStr1 failed ")
	}

	rm_rulePackage := 0
	if p.RmRulepackage {
		rm_rulePackage = 1
	}

	is_sql_exec := 0
	if p.IsSqlExec {
		is_sql_exec = 1
	}

	scheduledTime, _ := time.Parse(time.RFC3339, p.ScheduledTime)
	utcScheduledTime := scheduledTime.UTC()
	sqlStr2 := "insert into release_operations(task_id,host,rm_rulepackage,pkg_name,scheduled_time,is_sql_exec,canary_status) values (?,?,?,?,?,?,?)"
	ret2, err := tx.Exec(sqlStr2, p.TaskID, p.Host, rm_rulePackage, p.PkgName, utcScheduledTime, is_sql_exec, p.CanaryStatus)
	if err != nil {
		return
	}
	insertID, err := ret2.LastInsertId()
	if err != nil {
		return
	}
	logger.SugarLog.Infof("insert into release_operations row success,id:%v", insertID)
	return

}
