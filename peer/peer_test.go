package peer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeerParsing(t *testing.T) {
	for _, tt := range []struct {
		url  string
		peer Peer
	}{
		{
			url: "peer://staging@facade/facade-abcd",
			peer: Peer{
				InstanceName: "facade-abcd",
				AppName:      "facade",
				Environment:  "staging",
				ProjectName:  "facade",
			},
		},
		{
			url: "peer://facade/facade-abcd?project-name=facade-project",
			peer: Peer{
				InstanceName: "facade-abcd",
				AppName:      "facade",
				ProjectName:  "facade-project",
			},
		},
	} {
		p, err := ParsePeerURL(tt.url)

		assert.Nil(t, err)
		assert.Equal(t, p, &tt.peer)
		assert.Equal(t, tt.peer.URL().String(), tt.url)
	}
}
