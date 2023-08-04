package controller

type ParamsAccessKey struct {
	AccessKeyId     string `json:"access_key_id" form:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret" form:"access_key_secret"`
}

type ParamsTask struct {
	ParamsAccessKey
	TaskId string `json:"task_id" form:"task_id"`
}

type ParamsKubeConf struct {
	ParamsAccessKey
	ClusterId string `json:"cluster_id" form:"cluster_id" binding:"required"`
}

type ParamsSlbId struct {
	ParamsAccessKey
	ClusterId string `json:"cluster_id" form:"cluster_id" binding:"required"`
}

type ParamsDomainRecord struct {
	ParamsAccessKey
	Records []string `json:"records" binding:"required"`
	Value   string   `json:"value" binding:"required"`
}

type ParamUpload struct {
	LocalFilePath string `json:"local_file_path"`
	RemoteDir     string `json:"remote_dir"`
	RemoteHost    string `json:"host"`
}
