package discovery

import (
	"testing"
	"os"
)

func TestDiscoveryNormal(t *testing.T) {
	dir, err := os.Getwd()
	if err!=nil {
		t.Error(err)
	}
	t.Log("Working dir now:", dir)
}
