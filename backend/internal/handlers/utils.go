package handlers

import "strings"

// joinString joins a slice of strings with a separator
func joinString(ss []string, sep string) string {
	if len(ss) == 0 {
		return ""
	}
	if len(ss) == 1 {
		return ss[0]
	}
	n := len(sep) * (len(ss) - 1)
	for i := 0; i < len(ss); i++ {
		n += len(ss[i])
	}

	var b strings.Builder
	b.Grow(n)
	b.WriteString(ss[0])
	for _, s := range ss[1:] {
		b.WriteString(sep)
		b.WriteString(s)
	}
	return b.String()
}
