package pgmngr

import (
	"github.com/ParaServices/errgo"
)

// NewError ...
func NewError(e error) error {
	err, ok := e.(*errgo.Error)
	if ok {
		return err
	}
	return errgo.New(e)
}
