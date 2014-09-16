package torrent

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	ln, err := net.Listen("tcp", ":6881")
	defer ln.Close()
	assert.Nil(err)
	go func() {
		ln.Accept()
	}()
	conn, err := Connect(":6881")
	defer conn.Close()
	assert.Nil(err)
}

func TestHandshake(t *testing.T) {
	assert := assert.New(t)
	done := make(chan bool)
	ln, err := net.Listen("tcp", ":8080")
	defer ln.Close()
	if err != nil {
		t.Fatalf("coudln't start listener")
	}
	go func() {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			t.Fatalf("couldn't accept connection", err)
		}
		msg := fmt.Sprintf("%s%s", string(19), "BitTorrent protocol00000000abcdefghijklmnopqrstabcdefghijklmnopqrst")
		conn.Write([]byte(msg))
		data := make([]byte, 128)
		conn.Read(data)
		if string(data[0:68]) != msg {
			t.Fatalf(string(data) + " != " + msg)
		}
		done <- true
	}()
	conn, err := Connect(":8080")
	defer conn.Close()
	if err != nil {
		t.Fatalf("couldn't connect", err)
	}
	err = Handshake(conn, "abcdefghijklmnopqrst", "abcdefghijklmnopqrst")
	assert.Nil(err)
	<-done
}

func testSendMessage(t *testing.T, msg Message, expect []byte) {
	assert := assert.New(t)
	done := make(chan bool)
	ln, err := net.Listen("tcp", ":8080")
	defer ln.Close()
	if err != nil {
		t.Fatalf("coudln't start listener")
	}
	go func() {
		conn, err := ln.Accept()
		defer conn.Close()
		if err != nil {
			t.Fatalf("couldn't accept connection", err)
		}
		data := make([]byte, 32)
		conn.Read(data)
		assert.Equal(expect, data[0:len(expect)])
		done <- true
	}()
	conn, err := Connect(":8080")
	defer conn.Close()
	if err != nil {
		t.Fatalf("couldn't connect", err)
	}
	SendMessage(conn, msg)
	<-done
}

func TestSendMessage(t *testing.T) {
	testSendMessage(t, KeepAlive{}, []byte{0, 0, 0, 0})
	testSendMessage(t, Choke{}, []byte{0, 0, 0, 1, 0})
	testSendMessage(t, Unchoke{}, []byte{0, 0, 0, 1, 1})
	testSendMessage(t, Interested{}, []byte{0, 0, 0, 1, 2})
	testSendMessage(t, NotInterested{}, []byte{0, 0, 0, 1, 3})
	testSendMessage(t, Have{PieceIndex: 123456}, []byte{0, 0, 0, 5, 4, 0x0, 0x01, 0xe2, 0x40})
	testSendMessage(t, Bitfield{Data: []byte{1, 2, 3, 4}}, []byte{0, 0, 0, 5, 5, 1, 2, 3, 4})
	testSendMessage(t, Request{Index: 1, Begin: 2, Length: 3}, []byte{0, 0, 0, 13, 6, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3})
	testSendMessage(t, Piece{Index: 1, Begin: 2, Block: []byte{3, 4, 5}}, []byte{0, 0, 0, 12, 7, 0, 0, 0, 1, 0, 0, 0, 2, 3, 4, 5})
	testSendMessage(t, Cancel{Index: 1, Begin: 2, Length: 3}, []byte{0, 0, 0, 13, 8, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3})
}

func TestReceiveMessage(t *testing.T) {
	assert := assert.New(t)
	res := ParseMessage([]byte{0, 0, 0, 0})
	assert.Equal(res.(KeepAlive), KeepAlive{})
	res = ParseMessage([]byte{0, 0, 0, 1, 0})
	assert.Equal(res.(Choke), Choke{})
	res = ParseMessage([]byte{0, 0, 0, 1, 1})
	assert.Equal(res.(Unchoke), Unchoke{})
	res = ParseMessage([]byte{0, 0, 0, 1, 2})
	assert.Equal(res.(Interested), Interested{})
	res = ParseMessage([]byte{0, 0, 0, 1, 3})
	assert.Equal(res.(NotInterested), NotInterested{})
	res = ParseMessage([]byte{0, 0, 0, 5, 4, 0x0, 0x01, 0xe2, 0x40})
	assert.Equal(res.(Have), Have{PieceIndex: 123456})
	res = ParseMessage([]byte{0, 0, 0, 5, 5, 1, 2, 3, 4})
	assert.Equal(res.(Bitfield), Bitfield{Data: []byte{1, 2, 3, 4}})
	res = ParseMessage([]byte{0, 0, 0, 13, 6, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3})
	assert.Equal(res.(Request), Request{Index: 1, Begin: 2, Length: 3})
	res = ParseMessage([]byte{0, 0, 0, 12, 7, 0, 0, 0, 1, 0, 0, 0, 2, 3, 4, 5})
	assert.Equal(res.(Piece), Piece{Index: 1, Begin: 2, Block: []byte{3, 4, 5}})
	res = ParseMessage([]byte{0, 0, 0, 13, 8, 0, 0, 0, 1, 0, 0, 0, 2, 0, 0, 0, 3})
	assert.Equal(res.(Cancel), Cancel{Index: 1, Begin: 2, Length: 3})
}
