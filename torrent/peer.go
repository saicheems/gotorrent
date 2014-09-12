package torrent

import (
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	connectTimeout = 5
	pStr           = "BitTorrent protocol"

	KeepAlive = iota
	Choke
	Unchoke
	Interested
	NotInterested
	Have
	Bitfield
	Request
	Piece
	Cancel
)

// Connect returns a connection to the peer at raddr.
func Connect(raddr string) (net.Conn, error) {
	// Remote address.
	conn, err := net.DialTimeout("tcp", raddr, connectTimeout*time.Second)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// Handshake completes a handshake with a peer. It returns an error if it is not successful in any
// part of the process.
func Handshake(conn net.Conn, infoHash string, peerID string) error {
	err := sendHandshake(conn, infoHash, peerID)
	if err != nil {
		return err
	}
	err = receiveHandshake(conn, infoHash)
	if err != nil {
		return err
	}
	return nil
}

func SendMessage(conn net.Conn, messageType int, payload []byte) {

}

func ReadMessage(data []byte) (int, []byte) {
	return 0, nil
}

func sendHandshake(conn net.Conn, infoHash string, peerID string) error {
	magic := fmt.Sprintf("%s%s", string(len(pStr)), pStr)
	msg := fmt.Sprintf("%s%s%s%s", magic, "00000000", infoHash, peerID)
	_, err := conn.Write([]byte(msg))
	if err != nil {
		return err
	}
	return nil
}

func receiveHandshake(conn net.Conn, infoHash string) error {
	reply := make([]byte, 128)
	_, err := conn.Read(reply)
	if err != nil {
		return err
	}
	if reply[0] != byte(len(pStr)) {
		return errors.New("received pstr not expected length")
	}
	if string(reply[1:20]) != pStr {
		return errors.New("received pstr incorrect")
	}
	if string(reply[28:48]) != infoHash {
		return errors.New("received info_hash incorrect")
	}
	// reply[48:68] is the peerID of the connected peer.
	return nil
}
