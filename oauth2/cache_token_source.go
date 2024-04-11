package oauth2

import (
	"encoding/json"
	"os"

	"github.com/upfluence/errors"
	"golang.org/x/oauth2"

	"github.com/upfluence/pkg/log"
)

type CacheTokenSource struct {
	Filename    string
	TokenSource oauth2.TokenSource
}

func (cts *CacheTokenSource) Token() (*oauth2.Token, error) {
	t, err := cts.readToken()

	if err != nil || (t != nil && t.Valid()) {
		return t, errors.Wrap(err, "cant read the token")
	}

	t, err = cts.TokenSource.Token()

	if err != nil {
		return nil, errors.Wrap(err, "cant fetch underlying token")
	}

	if errw := cts.writeToken(t); errw != nil {
		log.WithError(err).Errorf("cant write token to %q", cts.Filename)
	}

	return t, nil
}

func (cts *CacheTokenSource) writeToken(t *oauth2.Token) error {
	buf, err := json.Marshal(t)

	if err != nil {
		return errors.Wrap(err, "cant marshal token")
	}

	return os.WriteFile(cts.Filename, buf, 0644)
}

func (cts *CacheTokenSource) readToken() (*oauth2.Token, error) {
	buf, err := os.ReadFile(cts.Filename)

	switch {
	case err == nil:
	case os.IsNotExist(err):
		return nil, nil
	default:
		return nil, err
	}

	var t oauth2.Token

	return &t, errors.Wrap(json.Unmarshal(buf, &t), "cant unmarshal token")
}
