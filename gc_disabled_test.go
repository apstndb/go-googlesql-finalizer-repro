//go:build disable_gc

package googlesqlrepro

func init() {
	disableGC = true
}
