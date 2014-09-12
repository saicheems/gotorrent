package torrent

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnnounce(t *testing.T) {
	tests := map[string]*AnnounceResponse{
		"d8:intervali1800e10:tracker id4:test8:completei1234e10:incompletei5678e5:peers4:abcde": &AnnounceResponse{Interval: 1800, TrackerID: "test", Complete: 1234, Incomplete: 5678, Peers: "abcd"},
		"d14:failure reason4:teste":                                                             &AnnounceResponse{FailureReason: "test"},
		"d15:warning message4:teste":                                                            &AnnounceResponse{WarningMessage: "test"},
	}

	for key, val := range tests {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, key)
		}))
		defer ts.Close()

		res, err := Announce(ts.URL)
		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(res, val)
	}
}

func TestGetAnnounceURL(t *testing.T) {
	tor := &Torrent{PeerID: "test", LocalPort: ":6881", Event: "started", MetaInfo: &MetaInfo{InfoHash: "test"}, Downloaded: 1234, Uploaded: 1234, Left: 1234, AnnounceURL: "http://test.com/announce"}
	url := tor.GetAnnounceURL()
	assert := assert.New(t)
	assert.Equal(url, "http://test.com/announce?compact=1&downloaded=1234&event=started&info_hash=test&left=1234&numwant=5&peer_id=test&port=6881&uploaded=1234")
}

func TestGetPeerAddresses(t *testing.T) {
	assert := assert.New(t)
	annResp := new(AnnounceResponse)
	annResp.Peers = "abcdefghijkl"
	assert.Equal(annResp.PeerAddresses(), []string{"97.98.99.100:25958", "103.104.105.106:27500"})
}
