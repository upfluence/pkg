package oauth2

import "golang.org/x/oauth2"

type ErrorTokenSource struct {
	Error error
}

func (ets ErrorTokenSource) Token() (*oauth2.Token, error) {
	return nil, ets.Error
}
