package config

type App struct {
	Env     string `mapstructure:"env" json:"env" yaml:"env"`
	AppName string `mapstructure:"app_name" json:"app_name" yaml:"app_name"`
	License string `mapstructure:"license" json:"license" yaml:"license"`
	Support string `mapstructure:"support" json:"support" yaml:"support"`
	Group   string `mapstructure:"group" json:"group" yaml:"group"`
}
