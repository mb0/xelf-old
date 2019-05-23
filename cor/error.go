package cor

import "golang.org/x/xerrors"

// centralize access to error api that is going to change in go 1.13
var (
	Error  = xerrors.New
	Errorf = xerrors.Errorf
)

// StrError returns a simple comparable error value that does not include stack information.
// This is usefully for lean package level error values.
func StrError(str string) error { return strError(str) }

type strError string

func (s strError) Error() string { return string(s) }

func IsErr(err, e error) bool { return xerrors.Is(err, e) }
