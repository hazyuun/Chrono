package config

import (
	"log"

	"github.com/spf13/viper"
)

type CfgPeriodic struct {
	Period int      `mapstructure:"period"`
	Files  []string `mapstructure:"files"`
}

type CfgSave struct {
	Files []string `mapstructure:"files"`
}

type CfgEvents struct {
	Periodic *CfgPeriodic `mapstructure:"periodic"`
	Save     *CfgSave     `mapstructure:"save"`
}

type CfgRoot struct {
	Events *CfgEvents `mapstructure:"events"`
}

var Cfg CfgRoot

func Load() {
	viper.SetConfigName("chrono")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil {
		log.Fatalf("Fatal error: couldn't load config file: %v", err.Error())
	}

	err = viper.Unmarshal(&Cfg)

	if err != nil {
		log.Fatalf("Fatal error: %v", err.Error())
	}
}
