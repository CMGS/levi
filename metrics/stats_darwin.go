package metrics

import (
	"../logs"
)

func NetNsSynchronize(pid string, fn func() error) (err error) {
	logs.Info("OSX not support Setns.")
	return fn()
}
