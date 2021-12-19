package aimerror

import "github.com/pkg/errors"

func FetchingUser(err error, username string) error {
	return errors.Wrapf(err, "could not fetch user with username %s", username)
}

func UserNotFound(username string) error {
	return errors.Errorf("no user with UIN %s", username)
}

var NoUserInSession = errors.New("no user in session")
