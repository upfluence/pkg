package sql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/upfluence/base/provider/credential"
)

func TestNullableThrift(t *testing.T) {
	for _, tt := range []struct {
		name      string
		wantValue *credential.CredentialReference
	}{
		{
			name:      "nil value",
			wantValue: nil,
		},
		{
			name: "with value",
			wantValue: &credential.CredentialReference{
				Type: credential.CredentialType_StripeConnectedAccount,
				Id:   42,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var (
				s = NullThrift[
					credential.CredentialReference,
					*credential.CredentialReference,
				]{
					Data: tt.wantValue,
				}

				data, err = s.Value()
			)

			require.NoError(t, err)

			s.Data = nil

			require.NoError(t, s.Scan(data))
			assert.Equal(t, tt.wantValue, s.Data)
		})
	}
}
