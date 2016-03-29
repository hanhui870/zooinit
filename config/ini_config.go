package config

import (
	"log"
	"os"
	"path"

	"github.com/codegangsta/cli"
	"github.com/go-ini/ini"
)

func Ini(path string) *ini.File {
	cfg, err := ini.Load(path)
	if err != nil {
		log.Fatalln(err)
	}

	ReloadWorkDir(cfg, "system")

	return cfg
}

func ReloadWorkDir(cfg *ini.File, section string) error {
	sec, err := cfg.GetSection(section)
	if err != nil {
		return err
	}

	// if pwd last part is bin, will use parent
	dirKey := sec.Key("work.dir")
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if path.Base(wd) == "bin" {
		wd = path.Dir(wd)
	}

	dirKey.SetValue(wd)

	return nil
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
