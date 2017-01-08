package conf

import "github.com/spf13/viper"


var Config *viper.Viper

func init() {
	Config = viper.New()
	Config.SetConfigName("kubrik")
	Config.AddConfigPath("/etc/kubrik")
	Config.AddConfigPath(".")
	//TODO: check error
	Config.ReadInConfig()
}