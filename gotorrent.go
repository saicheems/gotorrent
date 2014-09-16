package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/codegangsta/cli"
	"github.com/saicheems/gotorrent/torrent"
)

const (
	maxSeekConnections = 1
	maxConnections     = 55
)

func main() {
	app := cli.NewApp()
	app.Name = "gotorrent"
	app.Usage = "a minimal golang bittorrent client"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "port",
			Value: ":6881",
			Usage: "port for incoming connections",
		},
	}
	app.Action = func(c *cli.Context) {
		if len(c.Args()) != 1 {
			fmt.Println("one argument is required - a filepath to a .torrent file")
		} else {
			port := c.String("port")
			filePath := c.Args()[0]
			Start(port, filePath)
		}
	}
	app.Run(os.Args)
}

func Start(localPort string, filePath string) error {
	// Parse torrent and get Torrent struct.
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	t, err := torrent.New(generatePeerID(), localPort, f)
	if err != nil {
		return err
	}
	incomingAddresses := make(chan string)
	go Announcer(t, incomingAddresses)
	go PeerManager(t, localPort, incomingAddresses)
	fmt.Scanf("\n")
	return nil
}

// PeerManager starts a service that connects to peers as they come in and spins up peer handling
// threads. If we're connected to the maximum number of peers configured, the service will reject
// or close incoming connections.
func PeerManager(t *torrent.Torrent, localPort string, incomingAddresses chan string) error {
	totalConnections := 0
	peerQuit := make(chan bool) // Channel peers signal on when they die.
	incomingConnections := make(chan net.Conn)

	ln, err := net.Listen("tcp", localPort)
	if err != nil {
		return err
	}
	go func(incomingConnections chan net.Conn) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				continue
			}
			incomingConnections <- conn
		}
	}(incomingConnections)
	for {
		select {
		case <-peerQuit:
			totalConnections--
		default:
		}
		select {
		case in := <-incomingConnections:
			if totalConnections < maxConnections {
				go Talker(t, in, peerQuit)
				totalConnections++
			} else {
				conn := <-incomingConnections
				conn.Close()
			}
		default:
		}
		select {
		case in := <-incomingAddresses:
			if totalConnections < maxSeekConnections {
				conn, err := torrent.Connect(in)
				fmt.Println("Got incoming address...", err)
				if err == nil {
					go Talker(t, conn, peerQuit)
					totalConnections++
				}
			}
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func Talker(t *torrent.Torrent, conn net.Conn, peerQuit chan bool) {
	fmt.Println("Talking to peer.")
	err := torrent.Handshake(conn, t.MetaInfo.InfoHash, t.PeerID)
	if err == nil {
		for {
			// Do work.
			msg, err := torrent.ReadMessage(conn)
			if err != nil {
				continue
			}
			if msg != nil {
				fmt.Println(reflect.TypeOf(msg))
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	conn.Close()
	peerQuit <- true
	fmt.Println("Quit peer.")
}

func Announcer(t *torrent.Torrent, incomingConnections chan string) {
	for {
		annResp, err := torrent.Announce(t.GetAnnounceURL())
		if err != nil {
			continue
		}
		for _, addr := range annResp.PeerAddresses() {
			fmt.Println(addr)
			incomingConnections <- addr
		}
		time.Sleep(20 * time.Second)
	}
}

func generatePeerID() string {
	peerId := make([]byte, 20)
	rand.Read(peerId)
	return string(peerId)
}
