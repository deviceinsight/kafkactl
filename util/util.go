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

func StrArrayEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
