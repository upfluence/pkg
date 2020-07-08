package static

import (
	"bytes"
	"encoding/json"

	"github.com/upfluence/cfg/provider"
	pjson "github.com/upfluence/cfg/provider/json"
)

func NewProvider(d interface{}) provider.Provider {
	var (
		buf bytes.Buffer

		enc = json.NewEncoder(&buf)
	)

	if err := enc.Encode(d); err != nil {
		return provider.ProvideError("static", err)
	}

	return pjson.NewProviderFromReader(&buf)
}
