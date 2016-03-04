package cluster

import (
	"testing"

	"zooinit/bootstrap"
)

func TestENVInterfaceNormal(t *testing.T) {
	var env Env
	env = bootstrap.NewEnvInfoFile("../config/config.ini")

	t.Log("ENV result:", env.LocalIP(), env.Service())

	if env.Service() != "bootstrap" {
		t.Error("ENV result error:", env.Service())
	}
}

func TestClusterENVNormal(t *testing.T) {
	env = NewEnvInfoFile("../config/config.ini", "consul", "consul")

	t.Logf("Parsed env info values: %q", env)

	if env.service != "consul" {
		t.Error("Parsed env env.service error:", env.service)
	}
}
