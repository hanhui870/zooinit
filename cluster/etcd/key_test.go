package etcd

import (
	"testing"
)

func TestCreateInOrderKeyFormatTest(t *testing.T) {
	path := "/zooinit/discovery/cluster/consul/election/00000000000000000115"

	//invalid slice index -1 (index must be non-negative)
	//path= path[-1:] // error index
	if !CheckInOrderKeyFormat(path) {
		t.Error("found error CheckInOrderKeyFormat(path) for:", path)
	}

	if GetInOrderKeyValue(path) != 115 {
		t.Error("found error GetInOrderKeyValue(path) value for:", GetInOrderKeyValue(path))
	}

	t.Log("TestCreateInOrderKeyFormatTest finished.")
}
