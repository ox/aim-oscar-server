package aimerror

import "github.com/pkg/errors"

func UserNotFound(screenname string) error {
	return errors.Errorf("no user with UIN %s", screenname)
}

var NoUserInSession = errors.New("no user in session")
