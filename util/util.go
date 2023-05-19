package util

func IsBinary(contents []byte) bool {
	for _, ch := range contents {
		if ch == 0 {
			return true
		}
	}
	return false
}
