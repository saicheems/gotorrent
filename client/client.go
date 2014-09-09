package client

import (
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
)

// Client contains the information necessary to maintain the torrent client.
type Client struct {
	PeerID    string
	LocalPort string
}

// Torrent contains the state information of a torrent.
type Torrent struct {
	Client      *Client // Parent client.
	AnnounceURL string
	Event       string
	InfoHash    string
	Uploaded    int64
	Downloaded  int64
	Left        int64
	MetaInfo    *MetaInfo
	Peers       map[net.Conn]string
}

// New returns a new initialized Client object with a randomly generated PeerID.
func New(localPort string) *Client {
	c := new(Client)
	c.PeerID = generatePeerID()
	c.LocalPort = localPort
	return c
}

// NewTorrent returns an intialized torrent object. It holds a reference to the Client object which
// it spawned from in order to have the peerID/localPort.
func (c *Client) NewTorrent(r io.Reader) (*Torrent, error) {
	m, err := Parse(r)
	if err != nil {
		return nil, err
	}
	t := new(Torrent)
	t.Client = c
	t.AnnounceURL = m.Announce
	t.Event = "started"
	t.InfoHash = m.InfoHash
	t.MetaInfo = m
	return t, nil
}

// GetAnnounceURL returns the url to query the tracker for an announce with all parameters set
// according to the passed client and torrent structs.
func (t *Torrent) GetAnnounceURL() string {
	v := url.Values{}
	v.Add("peer_id", t.Client.PeerID)
	v.Add("port", t.Client.LocalPort[1:]) // TODO: This is a bad solution...
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

func generatePeerID() string {
	peerId := make([]byte, 20)
	rand.Read(peerId)
	return string(peerId)
}
