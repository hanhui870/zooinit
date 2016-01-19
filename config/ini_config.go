package config

import (
	"github.com/go-ini/ini"
	"log"
)

func Ini(path string)(*ini.File){
	cfg, err := ini.Load(path)
	if err!=nil {
		log.Panic(err)
	}

	return cfg
}