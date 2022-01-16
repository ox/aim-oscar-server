package aimerror

import "github.com/pkg/errors"

func FetchingUser(err error, screen_name string) error {
	return errors.Wrapf(err, "could not fetch user with screen_name %s", screen_name)
}

func UserNotFound(screen_name string) error {
	return errors.Errorf("no user with UIN %s", screen_name)
}

var NoUserInSession = errors.New("no user in session")
