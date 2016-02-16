package utility

import (
	"reflect"
	"strings"
	"testing"
)

func TestSliceNormalTest(t *testing.T) {
	list := []string{"1", "333", "444", "1", "555", "333", "1", "555"}

	//slice_test.go:14: Found RemoveDuplicateUnOrder: [555 1 333 444]
	//slice_test.go:18: Found RemoveDuplicateInOrder: [1 333 444 555]
	t.Log("TestSliceNormalTest orginal array:", list)
	unorder := RemoveDuplicateUnOrder(list)
	t.Log("Found RemoveDuplicateUnOrder:", unorder)

	order := RemoveDuplicateInOrder(list)
	t.Log("Found RemoveDuplicateInOrder:", order)
	if !reflect.DeepEqual(order, []string{"1", "333", "444", "555"}) {
		t.Error("Found RemoveDuplicateInOrder error:", order)
	}
}

func TestStringSlice(t *testing.T) {
	str := "TestParseCmdTestStringWithTestParams"

	var iter int
	find := "Test"
	//panic: runtime error: slice bounds out of range [recovered]
	//t.Log(find[5:])
	t.Log("String", str, "Slice based AT 5:", str[5:])

	start := 0
	for {
		chunk := str[start:]
		iter = strings.Index(chunk, find)
		if iter == -1 {
			t.Log("Find no", find, "in:", chunk)
			break
		} else {
			t.Log("Find", find, "in:", chunk, "At", iter)
			//iter是基于当前start的
			start = start + iter + len(find)
			if start > len(str) {
				t.Log("Final break At:", start)
				break
			}

			t.Log("Next round start At:", start)
		}
	}

	str = "fdsafda "
	t.Log("String trim of `fdsafda `:", strings.Trim(str, " fa"))
}
