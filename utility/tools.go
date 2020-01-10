package utility

// Check to see if a string exists in a list of strings
// https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func Contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
