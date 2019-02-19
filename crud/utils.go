package crud

import (
	"strings"
)

func getPrefix(url string) (prefix string) {
	if strings.Contains(prefix, "?") {
		return "&"
	} else {
		return "?"
	}
}
