package oauth2

import (
	"encoding/json"
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

type CacheTokenSource struct {
	Filename    string
	TokenSource oauth2.TokenSource
}

func (cts *CacheTokenSource) Token() (*oauth2.Token, error) {
	t, err := cts.readToken()

	if err != nil || (t != nil && t.Valid()) {
		return t, err
	}

	t, err = cts.TokenSource.Token()

	if err != nil {
		return nil, err
	}

	if errw := cts.writeToken(t); errw != nil {
		fmt.Fprintf(os.Stderr, "can not write token cache file: %v\n", errw)
	}

	return t, nil
}

func (cts *CacheTokenSource) writeToken(t *oauth2.Token) error {
	f, err := os.OpenFile(cts.Filename, os.O_RDWR|os.O_CREATE, 0644)

	if err != nil {
		return err
	}

	return json.NewEncoder(f).Encode(t)
}

func (cts *CacheTokenSource) readToken() (*oauth2.Token, error) {
	f, err := os.Open(cts.Filename)

	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	var t oauth2.Token

	return &t, json.NewDecoder(f).Decode(&t)
}
