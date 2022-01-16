package homeapp

import (
	"shanhu.io/homedrv/drvapi"
)

// Image returns the image of an app based on its meta information.
func Image(meta *drvapi.AppMeta) string {
	if meta.Image != "" {
		return meta.Image
	}
	if n := len(meta.Steps); n > 0 {
		last := meta.Steps[n-1]
		return last.Image
	}
	return ""
}
