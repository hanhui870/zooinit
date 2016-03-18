package cluster

import (
	"testing"
)

func TestClusterENVNormal(t *testing.T) {
	env = NewEnvInfoFile("../config/config.ini", "consul", "consul")

	t.Logf("Parsed env info values: %q", env)

	if env.service != "consul" {
		t.Error("Parsed env env.service error:", env.service)
	}
}
