package config

import (
	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"

	"log"
)

func Ini(path string) *ini.File {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalln(err)
	}

	return cfg
}

func GetValueString(keyNow string, sec *ini.Section, c *cli.Context) string {
	if c != nil && c.String(keyNow) != "" {
		return c.String(keyNow)
	} else {
		return sec.Key(keyNow).String()
	}
}

func GetValueInt(keyNow string, sec *ini.Section, c *cli.Context) (int, error) {
	if c != nil && c.Int(keyNow) != 0 {
		return c.Int(keyNow), nil
	} else {
		return sec.Key(keyNow).Int()
	}
}

func GetValueFloat64(keyNow string, sec *ini.Section, c *cli.Context) (float64, error) {
	if c != nil && c.Float64(keyNow) != 0 {
		return c.Float64(keyNow), nil
	} else {
		return sec.Key(keyNow).Float64()
	}
}
