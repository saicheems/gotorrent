package client

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
