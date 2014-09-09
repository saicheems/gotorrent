package client

import (
	"crypto/rand"
	"net"
)

// Torrent contains the state information of a torrent.
type Torrent struct {
	AnnounceURL string
	Event       string
	InfoHash    string
	Uploaded    int64
	Downloaded  int64
	Left        int64
	MetaInfo    *MetaInfo
	Peers       map[net.Conn]string
}

func GeneratePeerID() string {
	peerId := make([]byte, 20)
	rand.Read(peerId)
	return string(peerId)
}
