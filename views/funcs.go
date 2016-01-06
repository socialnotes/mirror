package views

import "fmt"

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
