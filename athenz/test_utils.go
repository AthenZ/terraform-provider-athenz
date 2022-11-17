package athenz

import (
	"fmt"
	"regexp"
)

func getPatternErrorRegex(attribute string) *regexp.Regexp {
	return getErrorRegex(fmt.Sprintf("%s must match the pattern", attribute))
}

func getErrorRegex(errorMessage string) *regexp.Regexp {
	r, _ := regexp.Compile(fmt.Sprintf("%s", errorMessage))
	return r
}
