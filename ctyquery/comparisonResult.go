package ctyquery

type ComparisonResult struct {
	ok      bool
	message string
	err     error
}

func newComparisonResult(ok bool, message string, err error) ComparisonResult {
	return ComparisonResult{
		ok:      ok,
		message: message,
		err:     err,
	}
}

func (m ComparisonResult) Err() error {
	return m.err
}

func (m ComparisonResult) Ok() bool {
	return m.ok
}

func (m ComparisonResult) Message() string {
	return m.message
}
