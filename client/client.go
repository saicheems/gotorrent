package client

import "fmt"
import "github.com/saicheems/gotorrent/client/announce"

// Client contains the information necessary for the client to run torrents.
type Client struct {
	peerID string
	port   string
}

// torrent contains the state information of a torrent.
type torrent struct {
}

// New returns a Client initialized using the .torrent file at filePath. If a problem is encountered
// while parsing the .torrent file, an error is returned.
func New(filePath string) (*Client, error) {
	c := new(Client)
	m, err := parse(filePath)
	fmt.Println(m)
	if err != nil {
		return nil, err
	}
	announce.Send("")
	return c, nil
}

// Start starts the torrent.
func (c *Client) Start() {

}

// Stop stops the torrent.
func (c *Client) Stop() {

}
