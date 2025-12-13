package filechk

import "os"

// IsFile determines if a file exists.
func IsFile(path string) bool {
	o, err := os.Stat(path)
	if err != nil {
		return false
	}
	if o.IsDir() == true {
		return false
	}
	return true
}

// IfNotEmptyAndIsFile returns false if a path is empty ("") else if the
// file exists.
func IfNotEmptyAndIsFile(path string) bool {
	if path == "" {
		return false
	}
	return IsFile(path)
}
