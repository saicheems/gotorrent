package torrent

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	connectTimeout = 5
	pStr           = "BitTorrent protocol"
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

// SendMessage writes the byte formatted Message to the provided connection.
func SendMessage(conn net.Conn, msg Message) {
	conn.Write(msg.Format())
}

func ReadMessage(conn net.Conn) (Message, error) {
	lenBuf := make([]byte, 4)
	_, err := conn.Read(lenBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[0:4])
	buf := make([]byte, length)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return ParseMessage(append(lenBuf, buf...)), nil
}

// ParseMessage returns a struct that implements Message containing all relevant information for a
// peer message. TODO: Currently returns nil if the parse fails. Should return some kind of error I
// think instead.
func ParseMessage(data []byte) Message {
	fmt.Println(data)
	// Can't possibly be good data if we don't even have a length.
	if len(data) < 4 {
		return nil
	}
	length := binary.BigEndian.Uint32(data[0:4])
	if length == 0 {
		return KeepAlive{}
	}
	// If the length isn't what it says it is, then fail.
	if length != uint32(len(data[4:])) {
		return nil
	}
	if length == 1 {
		switch data[4] {
		case 0: // choke
			return Choke{}
		case 1: // unchoke
			return Unchoke{}
		case 2: // interested
			return Interested{}
		case 3: // not interested
			return NotInterested{}
		}
	} else if length == 5 {
		if data[4] == 4 {
			pieceIndex := binary.BigEndian.Uint32(data[5:9])
			return Have{PieceIndex: pieceIndex}
		}
	} else if length == 13 {
		index := binary.BigEndian.Uint32(data[5:9])
		begin := binary.BigEndian.Uint32(data[9:13])
		length := binary.BigEndian.Uint32(data[13:17])
		if data[4] == 6 {
			return Request{Index: index, Begin: begin, Length: length}
		} else if data[4] == 8 {
			return Cancel{Index: index, Begin: begin, Length: length}
		}
	}
	if data[4] == 5 {
		return Bitfield{Data: data[5 : 5+length-1]}
	} else if data[4] == 7 {
		index := binary.BigEndian.Uint32(data[5:9])
		begin := binary.BigEndian.Uint32(data[9:13])
		return Piece{Index: index, Begin: begin, Block: data[13 : 13+length-9]}
	}
	return nil
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
	reply := make([]byte, 68)
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
