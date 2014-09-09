package main

import (
	"fmt"
	"os"

	"github.com/codegangsta/cli"
	"github.com/saicheems/gotorrent/client"
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
			start(c.String("port"), c.Args()[0])
		}
	}
	app.Run(os.Args)
}

func start(localPort string, filePath string) error {
	c := client.New(localPort)
	// Parse torrent and get Torrent struct.
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	m, err := client.Parse(f)
	if err != nil {
		return err
	}
	t := c.NewTorrent(m)
	fmt.Println(t)
	return nil
}
