package settings

import (
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"hkbackupCluster/logger"
)

var EnvMerchantMap map[string]string

func Init() (err error) {
	viper.SetConfigFile("./conf/merchantID.yaml")

	//读取初始配置文件
	err = viper.ReadInConfig()
	if err != nil {
		logger.SugarLog.Errorf("Failed to read file:%v", err)
		return
	}

	if err = viper.Unmarshal(&EnvMerchantMap); err != nil {
		logger.SugarLog.Errorf("viper Unmarshal failed,err:%v", err)
		return
	}

	//监听配置文变化并重新加载配置
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		logger.SugarLog.Info("Config is change,Reloading...")
		//重新加载配置
		err = viper.ReadInConfig()
		if err != nil {
			logger.SugarLog.Errorf("Failed to read file:%v", err)
			return
		}
	})
	return
}

func EnvFindMerchant(envNames []string) (merchantsID []string, invalidEnv []string) {
	for _, envName := range envNames {
		found := false
		for key, value := range EnvMerchantMap {
			if value == envName {
				found = true
				merchantsID = append(merchantsID, key)
				break
			}
		}
		if !found {
			invalidEnv = append(invalidEnv, envName)
		}
	}
	return
}

func MerchantFindEnv(merchantsID []string) (envNames []string) {
	for _, merchantID := range merchantsID {
		if envName, ok := EnvMerchantMap[merchantID]; ok {
			envNames = append(envNames, envName)
		}
	}
	return

}
