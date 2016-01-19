package config

import (
	"zooinit/log"
	"github.com/go-ini/ini"
)

func Ini(path string)(*ini.File){
	cfg, err := ini.Load(path)
	if err!=nil {
		log.Logger().Fatalln(err)
	}

	return cfg
}