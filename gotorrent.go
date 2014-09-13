package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"time"

	"github.com/codegangsta/cli"
	"github.com/saicheems/gotorrent/torrent"
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
	go Announcer(t)
	fmt.Scanf("\n")
	return nil
}

func Announcer(t *torrent.Torrent) {
	for {
		annResp, err := torrent.Announce(t.GetAnnounceURL())
		if err != nil {
			fmt.Println("Tracker announce failed")
			continue
		}
		addrs := annResp.PeerAddresses()
		for _, addr := range addrs {
			conn, err := torrent.Connect(addr)
			if err != nil {
				fmt.Println("Connection failed")
				continue
			}
			fmt.Println("Connection successful to peer", addr)
			err = torrent.Handshake(conn, t.InfoHash, t.PeerID)
			if err != nil {
				fmt.Println("Handshake failed")
				continue
			}
			fmt.Println("Handshake successful to peer", addr)
		}
		time.Sleep(120 * time.Second)
	}
}

func generatePeerID() string {
	peerId := make([]byte, 20)
	rand.Read(peerId)
	return string(peerId)
}
