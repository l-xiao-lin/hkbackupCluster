package model

type ParamsCompareJar struct {
	SrcHost  string `json:"src_host" form:"src_host"  binding:"required"`
	DestHost string `json:"dest_host" form:"dest_host" binding:"required"`
	AppName  string `json:"app_name"  form:"app_name" `
	SrcDir   string `json:"src_dir" form:"src_dir"`
	DestDir  string `json:"dest_dir" form:"dest_dir"`
}
