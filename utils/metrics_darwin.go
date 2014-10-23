package utils

func NetNsSynchronize(pid string, fn func() error) (err error) {
	Logger.Info("OSX not support Setns.")
	return fn()
}
