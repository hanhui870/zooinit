package config

import (
	"testing"
	"os"
)

func TestIniConfigNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Error(err)
	}

	file := dir + "/config_for_test.ini"
	t.Log("Working dir now:", dir, " Ini file:", file)

	ini := Ini(file)

	sec, err := ini.GetSection("production")
	if err != nil {
		t.Error(err)
	}

	key, err := sec.GetKey("log.path")
	if err != nil {
		t.Error(err)
	}else {
		if key.String() != "/Users/bruce/project/godev/src/zooinit/log" {
			t.Error("Get log.path error:", key.String())
		}else {
			t.Log("Get log.path:", key.String())
		}
	}

	exist := sec.HasKey("log")
	if exist {
		t.Error("Top section should not exist.")
	}

	secTest := ini.Section("production.testing")

	t.Log("Get log.ttl:", secTest.Key("log.ttl").String())
	t.Log("Get log.path:", ini.Section("production.testing").Key("log.path").String())

	for key, value := range secTest.KeysHash() {
		t.Log(key, value)
	}

	if key.String() != "/Users/bruce/project/godev/src/zooinit/log" {
		t.Error("Get log.path error for child testing:", key.String())
	}else {
		t.Log("Get log.path for child testing:", secTest.Key("log.path").String())
	}

	t.Log("Get array:", ini.Section("array").Key("FLOAT64S").Float64s(","))
	t.Log("Get time:", ini.Section("array").Key("TIMES").Times(","))
	t.Log("Get string quoted:", ini.Section("comments").Key("key3").Strings(","))

	t.Log([]string{"one", "two", "three"})
}


