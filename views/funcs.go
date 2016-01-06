package views

import "fmt"

func humanizeBytes(bytes int64) string {
	var (
		scale = []string{"b", "kb", "Mb", "Gb", "Tb", "Pb", "Eb", "Zb", "Yb"}
		rest  = bytes
		n     = 0
		s     = "mathisbroken"
	)
	for rest > 1000 {
		rest /= 1000
		n++
	}
	if n < len(scale) {
		s = scale[n]
	}

	return fmt.Sprintf("%d%s", rest, s)
}
