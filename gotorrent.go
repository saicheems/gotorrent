package main

import (
	"crypto/rand"
	"fmt"
	"net"
	"os"
	"reflect"
	"time"

	"github.com/codegangsta/cli"
	"github.com/saicheems/gotorrent/bitset"
	"github.com/saicheems/gotorrent/torrent"
)

const (
	maxSeekConnections = 1
	maxConnections     = 55
	keepAliveTimeout   = 110
	announcePeriod     = 200
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
	t, err := torrent.New(GeneratePeerID(), localPort, f)
	if err != nil {
		return err
	}

	fmt.Println(t.MetaInfo.Info.PieceLength)
	fmt.Println(t.MetaInfo.Info.Length)
	fmt.Println(len(t.MetaInfo.Info.Pieces) / 20)

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
				go Peer(t, in, peerQuit)
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
				fmt.Println("Got incoming address...", conn)
				if err == nil {
					go Peer(t, conn, peerQuit)
					totalConnections++
				}
			}
		default:
		}
		// Sleep for a bit so we don't hog up the goroutine scheduler.
		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// Peer starts a new Reader and Sender for a connection.
func Peer(t *torrent.Torrent, conn net.Conn, peerQuit chan bool) {
	go Reader(t, conn, peerQuit)
	go Sender(conn, make(chan torrent.Message))
}

func Reader(t *torrent.Torrent, conn net.Conn, peerQuit chan bool) {
	fmt.Println("Talking to peer.")
	err := torrent.Handshake(conn, t.MetaInfo.InfoHash, t.PeerID)
	if err == nil {
		l := t.MetaInfo.Info.Length / t.MetaInfo.Info.PieceLength
		bs := bitset.New(int(l))
		torrent.SendMessage(conn, &torrent.Bitfield{bs.Bytes()})
		for {
			// Deadline kills read with an error if we've waited too long without any
			// messages (2 minutes).
			conn.SetReadDeadline(time.Now().Add(time.Second * keepAliveTimeout))
			msg, err := torrent.ReadMessage(conn)
			if err != nil {
				fmt.Println("Killing...", err)
				break
			}
			if msg != nil {
				fmt.Println(reflect.TypeOf(msg))
			}
		}
	}
	conn.Close()
	peerQuit <- true
	fmt.Println("Quit peer.")
}

// Sender delivers messages that come in on the message channel. It also sends keep-alive messages
// periodically if a message hasn't come in for a fixed time period.
func Sender(conn net.Conn, msg chan torrent.Message) {
	for {
		select {
		case m := <-msg:
			conn.SetWriteDeadline(time.Now().Add(time.Second))
			err := torrent.SendMessage(conn, m)
			if err != nil {
				break
			}
		case <-time.After(time.Second * keepAliveTimeout):
			conn.SetWriteDeadline(time.Now().Add(time.Second))
			err := torrent.SendMessage(conn, &torrent.KeepAlive{})
			fmt.Println("Sending keep-alive.", err, err == nil)
			if err != nil {
				return
			}
		}
	}
}

// Announcer periodically announces to the tracker and pulls a new peer list. It passes this list to
// the peer manager.
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
		time.Sleep(time.Second * announcePeriod)
	}
}

// GeneratePeerID returns a 20 character random string to serve as the PeerID of the client.
func GeneratePeerID() string {
	peerId := make([]byte, 20)
	rand.Read(peerId)
	return string(peerId)
}
