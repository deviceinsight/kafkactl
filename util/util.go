package util

func Contains(list []string, element string) bool {
	for _, it := range list {
		if it == element {
			return true
		}
	}
	return false
}
