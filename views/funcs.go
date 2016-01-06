package views

import (
	"fmt"
	"strings"
)

func humanizeBytes(bytes int64) string {
	var (
		// int64 <= EB
		scale = []string{"B", "kB", "MB", "GB", "TB", "PB", "EB"}
		rest  = bytes
		n     = 0
	)
	for rest > 1024 {
		rest >>= 10
		n++
	}

	return fmt.Sprintf("%d%s", rest, scale[n])
}

func lazyEq(a, b interface{}) bool {
	as, aok := a.(string)
	bs, bok := b.(string)
	if aok && bok {
		return strings.Compare(as, bs) == 0
	} else {
		return a == b
	}
}
