package utils

import "strings"

func ParsePlace(place string) (string, string) {
	placeSplit := strings.SplitN(place, ".", 2)
	if len(placeSplit) == 1 {
		return "", placeSplit[0]
	}
	return placeSplit[0], placeSplit[1]
}
