package mysql

import (
	"hkbackupCluster/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func Init() (err error) {
	dsn := "root:jZ8tA4wX0jK@tcp(10.0.0.180:3306)/autoDeploy?charset=utf8mb4&parseTime=True"
	//dsn := "root:jZ8tA4wX0jK@tcp(10.0.0.180:3306)/demo?charset=utf8mb4&parseTime=True"
	db, err = sqlx.Connect("mysql", dsn)
	if err != nil {
		logger.SugarLog.Errorf("sqlx.Connect failed,err:%v", err)
		return
	}
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(20)
	return
}

func Close() {
	defer db.Close()
}
