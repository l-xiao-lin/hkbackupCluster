package model

import "github.com/gin-gonic/gin"

type Builder interface {
	Build(c *gin.Context) error
}

type ParamsEasyseller struct {
	JobName         string `json:"job_name" binding:"required"`
	SvnUrl          string `json:"SVNURL"`
	Hosts           string `json:"HOSTS"`
	ControllerConf  string `json:"CONTROLLERCONF"`
	ServiceConf     string `json:"SERVICECONF"`
	ConNamespace    string `json:"CONNAMESPACE"`
	ServiceGroup    string `json:"SERVICEGROUP"`
	ControllerGroup string `json:"CONTROLLERGROUP"`
}

func (p *ParamsEasyseller) Build(c *gin.Context) error {
	return c.ShouldBindJSON(p)
}

func NewParamsEasyseller() *ParamsEasyseller {
	return &ParamsEasyseller{
		JobName:         "test-easysellers",
		SvnUrl:          "http://10.0.0.109/ibiz/trunk/tongtool/src/BusinessFrontEnd/EasySellers/test/BackService",
		Hosts:           "testtwerp",
		ControllerConf:  "true",
		ServiceConf:     "true",
		ConNamespace:    "test-easysellers",
		ServiceGroup:    "twerp-service",
		ControllerGroup: "twerp-controller",
	}
}

type ParamsPackage struct {
	JobName           string `json:"job_name" binding:"required"`
	ProductLine       string `json:"Product_Line"`
	Project           string `json:"Project"`
	AppPackageUatHost string `json:"App_Package_Uat_Host"`
	TargetHost        string `json:"Target_Host"`
	AppName           string `json:"App_Name"`
	SvnProj           string `json:"svn_proj"`
	SvnDepth          string `json:"svn_depth"`
	Svn               string `json:"svn"`
	IsUpload          string `json:"Is_Upload"`
}

func (p *ParamsPackage) Build(c *gin.Context) error {
	return c.ShouldBindJSON(p)
}

func NewParamsPackage() *ParamsPackage {
	return &ParamsPackage{
		JobName:           "demo",
		ProductLine:       "业务前台",
		Project:           "product",
		AppPackageUatHost: "10.0.0.141",
		TargetHost:        "product_server",
		AppName:           "product-controller",
		SvnProj:           "ERP3.0",
		SvnDepth:          "1",
		Svn:               "版本模板/Patch/",
		IsUpload:          "true",
	}

}

type RespBuild struct {
	BuildNumber int64  `json:"build_number"`
	Result      string `json:"result"`
	ConsoleLog  string `json:"consoleLog,omitempty"`
}
