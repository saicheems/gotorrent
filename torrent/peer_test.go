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
	ln, err := net.Listen("tcp", ":8080")
	defer ln.Close()
	if err != nil {
		t.Fatalf("coudln't start listener")
	}
	go func() {
		//		time.Sleep(100)
		conn, err := ln.Accept()
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
	}()
	conn, err := Connect(":8080")
	defer conn.Close()
	if err != nil {
		t.Fatalf("couldn't connect", err)
	}
	err = Handshake(conn, "abcdefghijklmnopqrst", "abcdefghijklmnopqrst")
	assert.Nil(err)
}
