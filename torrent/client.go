package torrent

import (
	"io"
	"net"
)

// Torrent contains the state information of a torrent.
type Torrent struct {
	// Client related...
	PeerID    string
	LocalPort string
	// Torrent related...
	AnnounceURL string
	Event       string
	InfoHash    string
	Uploaded    int64
	Downloaded  int64
	Left        int64
	MetaInfo    *MetaInfo
	Peers       map[net.Conn]string
}

// NewTorrent returns an intialized torrent object. It holds a reference to the Client object which
// it spawned from in order to have the peerID/localPort.
func New(peerID string, localPort string, r io.Reader) (*Torrent, error) {
	m, err := Parse(r)
	if err != nil {
		return nil, err
	}
	t := new(Torrent)
	t.PeerID = peerID
	t.LocalPort = localPort
	t.AnnounceURL = m.Announce
	t.Event = "started"
	t.InfoHash = m.InfoHash
	t.MetaInfo = m
	return t, nil
}
