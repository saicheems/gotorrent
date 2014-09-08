package main

import (
	"os"

	"github.com/codegangsta/cli"
	"github.com/saicheems/gotorrent/client"
)

func main() {
	app := cli.NewApp()
	app.Name = "gotorrent"
	app.Usage = "bittorrent in golang"
	app.Action = func(c *cli.Context) {
		if len(c.Args()) > 0 {
			client.New(c.Args()[0])
		}
	}
	app.Run(os.Args)
}
