package evolution

import "strings"

func transformScopedName(pkg string) string {
	return strings.Replace(pkg, "/", "%2f", -1)
}
