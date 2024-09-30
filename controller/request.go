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

type ParamWeChat struct {
	Message string `json:"message"`
	CorpID  string `json:"corp_id"`
	Secret  string `json:"secret"`
	ToParty string `json:"toParty"`
	AgentID int    `json:"agent_id"`
}

type ParamCheck struct {
	EnvName string `json:"env_name" form:"env_name"`
}

type ParamListing struct {
	Website  string `json:"website" form:"website"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type ParamConf struct {
	Namespace string `json:"namespace" form:"namespace"`
	DataID    string `json:"dataID" form:"dataID"`
	Group     string `json:"group" form:"group"`
	Content   string `json:"content" form:"content"`
}

type ParamYms struct {
	Website  string `json:"website" form:"website"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}
