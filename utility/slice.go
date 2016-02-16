package utility

// Remove dulicates string in a slice keep order
func RemoveDuplicateInOrder(elements []string) []string {
	check := map[string]bool{}

	for v := range elements {
		check[elements[v]] = true
	}

	uniqueSlice := []string{}
	for key, value := range elements {
		if check[elements[key]] {
			check[elements[key]] = false
			uniqueSlice = append(uniqueSlice, value)
		}
	}
	return uniqueSlice
}

// Remove dulicates string in a slice unorder
func RemoveDuplicateUnOrder(elements []string) []string {
	check := map[string]bool{}

	for v := range elements {
		check[elements[v]] = true
	}

	uniqueSlice := []string{}
	for key, _ := range check {
		uniqueSlice = append(uniqueSlice, key)
	}
	return uniqueSlice
}
