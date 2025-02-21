package model

type ParamsUpdateStatus struct {
	TaskID string `json:"task_id" db:"task_id"  binding:"required"`
	Status int8   `json:"status" db:"status" binding:"required"`
}

type ParamsIncrementalPack struct {
	TaskID          string   `json:"task_id" db:"task_id"  binding:"required"`
	JobName         string   `json:"job_name" db:"job_name"  binding:"required"`
	Host            string   `json:"host"  db:"host" binding:"required"`
	Common          *string  `json:"common" db:"common" ` //可选字段，存表中变成null而不是空字符串
	Diff            *string  `json:"diff" db:"diff"`
	Status          int      `json:"status" db:"status"`
	SrcPath         string   `json:"src_path" db:"src_path"`
	RmRulepackage   bool     `json:"rm_rulepackage" db:"rm_rulepackage"`
	PkgName         *string  `json:"pkg_name" db:"pkg_name"`
	UpdateJbossConf bool     `json:"update_jbossconf" db:"update_jbossconf"`
	JbossConf       []Config `json:"jboss_conf" db:"-"`
	UpdateSdkConf   bool     `json:"update_sdkconf" db:"update_sdkconf"`
	SdkConf         []Config `json:"sdk_conf" db:"-"`
	UpdateSecurity  bool     `json:"update_security" db:"update_security"`
	IsSqlExec       bool     `json:"is_sql_exec" db:"is_sql_exec"`
	IsPackage       bool     `json:"is_package" db:"is_package"`
	CanaryStatus    *int     `json:"canary_status" db:"canary_status"` //1:需要灰度;2:取消灰度
	ScheduledTime   string   `json:"scheduled_time" db:"scheduled_time"`
	PackageTime     string   `json:"package_time" db:"package_time"`
	ShouldSend      bool     `json:"should_send" db:"should_send"`
}

type Config struct {
	ConfigType    string `json:"SELECT" db:"config_type" binding:"required"`
	ConfigContent string `json:"Config_Content" db:"config_content" binding:"required"`
	ConfigAction  string `json:"Action" db:"config_action" binding:"required"`
	Host          string `json:"Dest_hosts"  db:"host"`
}

type ParamTestPack struct {
	JobName       string `json:"job_name" binding:"required"`
	SrcPath       string `json:"src_path"`
	SrcHost       string `json:"src_host"`
	Host          string `json:"host"`
	Common        string `json:"common"`
	VersionOption string `json:"version_option"`
}

type RespPackageData struct {
	TaskID          string  `json:"task_id" db:"task_id"  binding:"required"`
	JobName         string  `json:"job_name" db:"job_name"  binding:"required"`
	Host            string  `json:"host"  db:"host" binding:"required"`
	Common          *string `json:"common" db:"common" ` //可选字段，存表中变成null而不是空字符串
	Diff            *string `json:"diff" db:"diff"`
	Status          int     `json:"status" db:"status"`
	SrcPath         string  `json:"src_path" db:"src_path"`
	RmRulepackage   bool    `json:"rm_rulepackage" db:"rm_rulepackage"`
	PkgName         *string `json:"pkg_name" db:"pkg_name"`
	UpdateJbossConf bool    `json:"update_jbossconf" db:"update_jbossconf"`
	UpdateSdkConf   bool    `json:"update_sdkconf" db:"update_sdkconf"`
	UpdateSecurity  bool    `json:"update_security" db:"update_security"`
	IsSqlExec       bool    `json:"is_sql_exec" db:"is_sql_exec"`
	IsPackage       bool    `json:"is_package" db:"is_package"`
	CanaryStatus    *int    `json:"canary_status" db:"canary_status"`
	ScheduledTime   string  `json:"scheduled_time" db:"scheduled_time"`
	PackageTime     string  `json:"package_time" db:"package_time"`
	ShouldSend      bool    `json:"should_send" db:"should_send"`
}

type ServiceStop struct {
	Host               string `json:"host"`
	RemoveJbossMonitor bool   `json:"remove_jboss_monitor"`
	RemoveSdkMonitor   bool   `json:"remove_sdk_monitor"`
	StopJboss          bool   `json:"stop_jboss"`
	StopSdk            bool   `json:"stop_sdk"`
	RestartJboss       bool   `json:"restart_jboss"`
	RestartSdk         bool   `json:"restart_sdk"`
	KeepMonitorJboss   bool   `json:"keep_monitor_jboss"`
	KeepMonitorSdk     bool   `json:"keep_monitor_sdk"`
}

type ParamReleaseXML struct {
	TaskID        string  `json:"task_id" db:"task_id"  binding:"required"`
	SrcPath       string  `json:"src_path" db:"src_path" binding:"required"`
	JobName       string  `json:"job_name" db:"job_name"  binding:"required"`
	Host          string  `json:"host"  db:"host" binding:"required"`
	Common        *string `json:"common" db:"common"`
	ScheduledTime string  `json:"scheduled_time" db:"scheduled_time"`
}
