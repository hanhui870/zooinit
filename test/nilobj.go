package main

type obj struct {
	test string
}

func (o obj) String() string {
	return o.test
}

func main() {
	var inst *obj

	inst = nil

	// panic: runtime error: invalid memory address or nil pointer dereference
	println(inst.String())
}
