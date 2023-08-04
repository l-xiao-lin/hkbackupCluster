package model

type WorkFlow struct {
	AccessKeyId     string   `json:"access_key_id" form:"access_key_id" binding:"required"`
	AccessKeySecret string   `json:"access_key_secret" form:"access_key_secret" binding:"required"`
	Records         []string `json:"records" form:"records" binding:"required"`
	LocalFilePath   string   `json:"local_file_path" binding:"required"`
	RemoteDir       string   `json:"remote_dir" binding:"required"`
	RemoteHost      string   `json:"host" binding:"required"`
}
