package model

type ParamCommonDeploy struct {
	TaskID           string         `json:"task_id" db:"task_id"`
	ServiceName      string         `json:"service_name" db:"service_name"`
	HasConfiguration bool           `json:"has_configuration" db:"has_configuration"`
	Status           int            `json:"status" db:"status"`
	OpenSchema       bool           `json:"open_schema" db:"open_schema"`
	CommonConfigs    []CommonConfig `json:"common_configs" db:"-"`
	PackageTime      string         `json:"package_time" db:"package_time"`
	ScheduledTime    string         `json:"scheduled_time" db:"scheduled_time"`
}

type CommonConfig struct {
	Action  string `json:"config_action" db:"config_action"`
	Content string `json:"config_content" db:"config_content"`
}

type RespPackageCommonData struct {
	TaskID           string `json:"task_id" db:"task_id"`
	ServiceName      string `json:"service_name" db:"service_name"`
	HasConfiguration bool   `json:"has_configuration" db:"has_configuration"`
	OpenSchema       bool   `json:"open_schema" db:"open_schema"`
	Status           int    `json:"status" db:"status"`
	PackageTime      string `json:"package_time" db:"package_time"`
	ScheduledTime    string `json:"scheduled_time" db:"scheduled_time"`
}

type RespConfCommon struct {
	TaskID      string `json:"task_id" db:"task_id"`
	ServiceName string `json:"service_name" db:"service_name"`
	Action      string `json:"config_action" db:"config_action"`
	Content     string `json:"config_content" db:"config_content"`
}
