package helpers

import "github.com/hashicorp/go-multierror"

func ReadErrors(errs chan error) error {
	var err error
	for e := range errs {
		err = multierror.Append(err, e)
	}
	return err
}
