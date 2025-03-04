package respond

// ErrorInfoer is an optional interface that errors can optionally implement to provide
// additional context in an error response.
type ErrorInfoer interface {
	error
	Info() any
}

type errWithInfo struct {
	err  error
	info any
}

func (ei errWithInfo) Error() string {
	return ei.err.Error()
}

func (ei errWithInfo) Info() any {
	return ei.info
}

func ErrWithInfo(err error, info any) errWithInfo {
	return errWithInfo{
		err:  err,
		info: info,
	}
}
