// Copyright 2016 祝景法(Bruce)@haimi.com. www.haimi.com All rights reserved.
package config

import (
	"os"
	"testing"
)

func TestIniConfigNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	file := dir + "/config_for_test.ini"
	t.Log("Working dir now:", dir, " Ini file:", file)

	ini := Ini(file)

	sys, err := ini.GetSection("system")
	if err != nil {
		t.Error(err)
	}

	key, err := sys.GetKey("work.dir")
	if err != nil {
		t.Error(err)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			t.Error("Error get work dir:", err)
		}
		if key.String() != wd {
			t.Error("Get work.dir error:", key.String())
		} else {
			t.Log("Get work.dir:", key.String())
		}
	}

	sec, err := ini.GetSection("system.production")
	if err != nil {
		t.Error(err)
	}

	key, err = sec.GetKey("log.path")
	if err != nil {
		t.Error(err)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			t.Error("Error get work dir:", err)
		}
		if key.String() != wd+"/log" {
			t.Error("Get log.path error:", key.String())
		} else {
			t.Log("Get log.path:", key.String())
		}
	}

	exist := sec.HasKey("log")
	if exist {
		t.Error("Top section should not exist.")
	}

	secTest := ini.Section("system.production.testing")

	if secTest.Key("log.ttl").String() != "10" {
		t.Log("Get log.ttl Error:", secTest.Key("log.ttl").String())
	} else {
		t.Log("Get log.ttl:", secTest.Key("log.ttl").String())
	}
	t.Log("Get log.path:", ini.Section("system.production.testing").Key("log.path").String())

	//loop test
	for key, value := range secTest.KeysHash() {
		t.Log(key, value)
	}

	t.Log("Get array:", ini.Section("array").Key("FLOAT64S").Float64s(","))
	t.Log("Get time:", ini.Section("array").Key("TIMES").Times(","))
	t.Log("Get string quoted:", ini.Section("comments").Key("key3").Strings(","))

	t.Log([]string{"one", "two", "three"})

	//test change dir
	os.Chdir("/usr/bin")
	ini = Ini(file)
	sec, err = ini.GetSection("system.production")
	if err != nil {
		t.Error(err)
	}

	key, err = sec.GetKey("log.path")
	if err != nil {
		t.Error(err)
	} else {
		if key.String() != "/usr/log" {
			t.Error("Get log.path error:", key.String())
		} else {
			t.Log("Get log.path:", key.String())
		}
	}
}
