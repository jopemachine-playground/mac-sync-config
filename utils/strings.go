package utils

import "strings"

// TODO: Replace below function with stdlib's one when it is merged
// Ref: https://stackoverflow.com/questions/10485743/contains-method-for-a-slice
func StringContains(slice []string, item string) bool {
	set := make(map[string]struct{}, len(slice))
	for _, s := range slice {
		set[s] = struct{}{}
	}

	_, ok := set[item]
	return ok
}

func SubStringContains(lstr string, rstr string) bool {
	return strings.Contains(strings.ToLower(lstr), strings.ToLower((rstr)))
}
