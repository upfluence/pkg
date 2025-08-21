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
			url: "peer://staging@k8s.upfluence.co/facade-abcd?app-name=facade",
			peer: Peer{
				Authority:    "k8s.upfluence.co",
				InstanceName: "facade-abcd",
				AppName:      "facade",
				Environment:  "staging",
				ProjectName:  "facade",
			},
		},
		{
			url: "peer://local/facade-abcd?project-name=facade-project",
			peer: Peer{
				Authority:    "local",
				InstanceName: "facade-abcd",
				AppName:      "facade-abcd",
				ProjectName:  "facade-project",
			},
		},
		{
			url: "peer://local/facade-abcd?buz=biz&buz=bar&project-name=facade-project",
			peer: Peer{
				Authority:    "local",
				InstanceName: "facade-abcd",
				AppName:      "facade-abcd",
				ProjectName:  "facade-project",
				Metadata: map[string][]string{
					"buz": {"biz", "bar"},
				},
			},
		},
	} {
		p, err := ParsePeerURL(tt.url)

		assert.Nil(t, err)
		assert.Equal(t, p, &tt.peer)
		assert.Equal(t, tt.peer.URL().String(), tt.url)
	}
}
