package torrent

import (
	"bytes"
	"fmt"
	"net"
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

// Announce sends an announce signal to a url and returns an AnnounceResponse. If there's a failure
// then the appropriate error is returned.
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

// GetAnnounceURL returns the url to query the tracker for an announce with all parameters set
// according to the Torrent.
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

// GetPeerAddresses returns a string list containing the formatted addresses of all peers in the
// AnnounceResponse.
func (response *AnnounceResponse) PeerAddresses() []string {
	peers := response.Peers
	addresses := make([]string, 0)
	for i := 0; i < len(peers); i += 6 {
		addresses = append(addresses, decodePeerAddress(peers[i:i+6]))
	}
	return addresses
}

// decodePeerAddress returns the readable address of a peer from its compact 6 byte chunk. The input
// string must be 6 bytes long.
func decodePeerAddress(chunk string) string {
	ip := net.IPv4(chunk[0], chunk[1], chunk[2], chunk[3])
	remotePort := 256*int(chunk[4]) + int(chunk[5]) // Port is given in network encoding.
	return fmt.Sprintf("%s:%d", ip.String(), remotePort)
}
