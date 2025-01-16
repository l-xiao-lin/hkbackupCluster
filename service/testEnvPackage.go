package service

import (
	"hkbackupCluster/logger"
	"hkbackupCluster/model"
	"hkbackupCluster/pkg/pack"
)

func TestEnvPackage(p *model.ParamTestPack) (resp *pack.RespBuild, err error) {
	//将参数转换成map
	paramMap, err := pack.StructToMap(p)
	if err != nil {
		logger.SugarLog.Errorf("pack.StructToMap failed,err:%v", err)
		return
	}

	resp, err = pack.JenkinsBuild(p.JobName, paramMap)
	if err != nil {
		return
	}

	return
}
