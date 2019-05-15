package util

func ContainsString(list []string, element string) bool {
	for _, it := range list {
		if it == element {
			return true
		}
	}
	return false
}

func ContainsInt32(list []int32, element int32) bool {
	for _, it := range list {
		if it == element {
			return true
		}
	}
	return false
}
