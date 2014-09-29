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
	maxSeekConnections = 5
	maxConnections     = 55
	keepAliveTimeout   = 110
	announcePeriod     = 20
)

type Client struct {
	LocalPort string
	Torrent   *torrent.Torrent
	OutFile   *os.File
	BitSet    *bitset.BitSet
}

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
	c := new(Client)
	c.LocalPort = localPort
	c.Torrent = t
	out, err := os.Create(t.MetaInfo.Info.Name)
	if err != nil {
		return err
	}

	c.OutFile = out
	b := bitset.New(int(t.MetaInfo.Info.Length / t.MetaInfo.Info.PieceLength))
	c.BitSet = b
	fmt.Println("Created file with", len(b.Bytes()), "pieces. Piece length:", t.MetaInfo.Info.PieceLength)
	incomingAddresses := make(chan string)
	incomingPieces := make(chan torrent.Piece, 256)
	outgoingRequests := make(chan torrent.Request, 256)
	go Announcer(t, incomingAddresses)
	go PeerManager(c, incomingAddresses, incomingPieces, outgoingRequests)
	go Writer(c, incomingPieces, outgoingRequests)
	fmt.Scanf("\n")
	return nil
}

// PeerManager starts a service that connects to peers as they come in and spins up peer handling
// threads. If we're connected to the maximum number of peers configured, the service will reject
// or close incoming connections.
func PeerManager(c *Client, incomingAddresses chan string, incomingPieces chan torrent.Piece, outgoingRequests chan torrent.Request) error {
	totalConnections := 0
	peerQuit := make(chan bool) // Channel peers signal on when they die.
	incomingConnections := make(chan net.Conn)

	ln, err := net.Listen("tcp", c.LocalPort)
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
				go Peer(c, in, incomingPieces, outgoingRequests, peerQuit)
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
					go Peer(c, conn, incomingPieces, outgoingRequests, peerQuit)
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

func Writer(c *Client, incomingPieces chan torrent.Piece, outgoingRequests chan torrent.Request) {
	pieceIndex := 0
	pieceLength := c.Torrent.MetaInfo.Info.PieceLength
	buf := make([]byte, pieceLength)
	bs := bitset.New(int(pieceLength / (1 << 14)))
	timeout := make([]time.Time, 32)
	for n := 0; n < 32; n++ {
		timeout[n] = time.Now()
	}
	for {
		select {
		case piece, ok := <-incomingPieces:
			if !ok {
				break
			}
			fmt.Println("Copied part of piece", pieceIndex, "at offset", piece.Begin)
			copy(buf[piece.Begin:int(piece.Begin)+len(piece.Block)], piece.Block)
			bs.Set(int(piece.Begin / (1 << 14)))
		default:
		}
		if bs.FirstZeroBit() < 0 {
			fmt.Println("Wrote piece", pieceIndex)
			c.BitSet.Set(pieceIndex)
			c.OutFile.WriteAt(buf, pieceLength*int64(pieceIndex))
			pieceIndex++
			bs = bitset.New(int(pieceLength / (1 << 14)))
			buf = make([]byte, pieceLength)
			for n := 0; n < 32; n++ {
				timeout[n] = time.Now()
			}
		} else {
			for n, t := range timeout {
				if !bs.Check(n) {
					if time.Now().After(t) {
						fmt.Println("New outgoing request... pieceIndex:", pieceIndex, "offset:", n*(1<<14), "length:", 1<<14)
						select {
						case outgoingRequests <- torrent.Request{uint32(pieceIndex), uint32(n * (1 << 14)), 1 << 14}:
						default:
						}
						timeout[n] = time.Now().Add(10 * time.Second)
					}
				}
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// Peer starts a new Reader and Sender for a connection.
func Peer(c *Client, conn net.Conn, incomingPieces chan torrent.Piece, outgoingRequests chan torrent.Request, peerQuit chan bool) {
	t := c.Torrent
	msgIn := make(chan torrent.Message)
	msgOut := make(chan torrent.Message)
	defer func() { peerQuit <- true; fmt.Println("Quit and closed peer.") }()
	defer conn.Close()
	err := torrent.Handshake(conn, t.MetaInfo.InfoHash, t.PeerID)
	if err != nil {
		return
	}
	go Reader(conn, msgIn)
	go Sender(conn, msgOut)
	msgOut <- torrent.Bitfield{c.BitSet.Bytes()}
	msgOut <- torrent.Interested{}
	choke := true
	for {
		select {
		case msg, ok := <-msgIn:
			if !ok {
				fmt.Println("Reader closed.")
				return
			} else {
				fmt.Println("Reading...", reflect.TypeOf(msg))
				switch m := msg.(type) {
				case torrent.Choke:
					choke = true
				case torrent.Unchoke:
					choke = false
				case torrent.Interested:
				case torrent.NotInterested:
				case torrent.Piece:
					// Send out the piece to the writer. Don't block.
					select {
					case incomingPieces <- m:
					default:
					}
				default:
				}
			}
		default:
		}
		if !choke {
			select {
			case m := <-outgoingRequests:
				fmt.Println("Added outgoing request to msg queue.")
				msgOut <- m
			default:
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("Quitting peer.")
}

func Reader(conn net.Conn, msgIn chan torrent.Message) {
	for {
		// Deadline kills read with an error if we've waited too long without any
		// messages (2 minutes).
		conn.SetReadDeadline(time.Now().Add(time.Second * keepAliveTimeout))
		msg, err := torrent.ReadMessage(conn)
		if err != nil {
			fmt.Println("Reading quitting:", err)
			close(msgIn)
			break
		}
		if msg != nil {
			msgIn <- msg
		}
	}
}

// Sender delivers messages that come in on the message channel. It also sends keep-alive messages
// periodically if a message hasn't come in for a fixed time period.
func Sender(conn net.Conn, msgOut chan torrent.Message) {
	for {
		var m torrent.Message
		select {
		case m = <-msgOut:
			fmt.Println("Sending...", reflect.TypeOf(m), m)
			err := torrent.SendMessage(conn, m)
			if err != nil {
				return
			}
			fmt.Println("Sent message")
		case <-time.After(time.Second * keepAliveTimeout):
			err := torrent.SendMessage(conn, torrent.KeepAlive{})
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
		fmt.Println("Announcing...")
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
