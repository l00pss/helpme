package goerr

type GoErr struct {
	error
	runtimeErr bool
}

func newGoErr(err error, isRuntime bool) *GoErr {
	return &GoErr{
		error:      err,
		runtimeErr: isRuntime,
	}
}

func (g *GoErr) IsRuntime() bool {
	return g.runtimeErr
}

func (g *GoErr) Unwrap() error {
	return g.error
}

func (g *GoErr) Error() string {
	return g.error.Error()
}

func WrapRuntimeErr(err error) *GoErr {
	return newGoErr(err, true)
}

func WrapNonRuntimeErr(err error) *GoErr {
	return newGoErr(err, false)
}

func IsGoErr(err error) bool {
	_, ok := err.(*GoErr)
	return ok
}
