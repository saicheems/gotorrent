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
	app.Run(os.Args)
	client.New("")
}
