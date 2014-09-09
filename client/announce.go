package client

import (
	"bytes"
	"net/http"

	bencode "github.com/jackpal/bencode-go"
)

// AnnounceResponse contains the information returned by a tracker request.
type AnnounceResponse struct {
	FailureReason  string "failure reason"
	WarningMessage string "warning message"
	Interval       int    "interval"
	MinInterval    int    "min interval"
	TrackerID      string "tracker id"
	Complete       int    "complete"
	Incomplete     int    "incomplete"
	Peers          string "peers"
}

// Announce sends an announce signal to a url and returns the response formatted to an
// AnnounceResponse.
func Announce(url string) (*AnnounceResponse, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	annRes := new(AnnounceResponse)
	err = bencode.Unmarshal(buf, annRes)
	if err != nil {
		return nil, err
	}
	return annRes, nil
}
