package cluster

import (
	"testing"

	"zooinit/bootstrap"
)

func TestENVInterfaceNormal(t *testing.T) {
	var env Env
	env =bootstrap.NewEnvInfoFile("../config/config.ini")

	t.Log("ENV result:", env.LocalIP(), env.Service())

	if env.Service() != "bootstrap" {
		t.Error("ENV result error:", env.Service())
	}
}