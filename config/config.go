package config

import (
	"log"
	"os"

	"github.com/go-ini/ini"

	"path/filepath"
)

func GetConfigFileName(configfile string) string {
	if configfile == "" {
		log.Fatalln("Bootstrap configFile path has not set.")
	}

	fname, err := filepath.Abs(configfile)
	if err != nil {
		log.Fatalln("Fetch Abs config file error:", err)
	}
	log.Println("Use configuration file:", fname)

	return fname
}

func GetConfigInstance(fname string) *ini.File {
	_, err := os.Stat(fname)
	if os.IsNotExist(err) {
		log.Fatalln("Configuration file:", err)
	}

	iniInstance := Ini(fname)
	return iniInstance
}
