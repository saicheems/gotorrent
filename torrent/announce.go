package torrent

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

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
func (t *Torrent) Announce() (*AnnounceResponse, error) {
	url := t.GetAnnounceURL()
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	fmt.Println(string(buf.Bytes()))
	annRes := new(AnnounceResponse)
	err = bencode.Unmarshal(buf, annRes)
	if err != nil {
		return nil, err
	}
	return annRes, nil
}

// GetAnnounceURL returns the url to query the tracker for an announce with all parameters set
// according to the passed client and torrent structs.
func (t *Torrent) GetAnnounceURL() string {
	v := url.Values{}
	v.Add("peer_id", t.PeerID)
	v.Add("port", t.LocalPort[1:]) // TODO: This is a bad solution...
	v.Add("event", t.Event)
	v.Add("info_hash", t.MetaInfo.InfoHash)
	// These are int64s so we have to use FormatInt.
	downloaded := strconv.FormatInt(t.Downloaded, 10)
	uploaded := strconv.FormatInt(t.Uploaded, 10)
	left := strconv.FormatInt(t.Left, 10)
	v.Add("downloaded", downloaded)
	v.Add("uploaded", uploaded)
	v.Add("left", left)
	v.Add("numwant", strconv.Itoa(5))
	v.Add("compact", "1") // We will always make compact requests.

	return fmt.Sprintf("%s?%s", t.AnnounceURL, v.Encode())
}
