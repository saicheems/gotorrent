package client

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"

	"github.com/jackpal/bencode-go"
)

// MalformedTorrentError is the error returned when a torrent file couldn't be parsed.
var MalformedTorrentError = fmt.Errorf("Malformed torrent file.")

// MetaInfo implements the contents and file structure of a .torrent file.
type metaInfo struct {
	Info         infoDict "info"
	InfoHash     string
	Announce     string     "announce"
	AnnounceList [][]string "announce-list"
	CreationDate string     "creation date"
	Comment      string     "comment"
	CreatedBy    string     "created by"
	Encoding     string     "encoding"
}

type infoDict struct {
	PieceLength int64      "piece length"
	Pieces      string     "pieces"
	Private     int64      "private"
	Name        string     "name"
	Length      int64      "length"
	Md5Sum      string     "md5sum"
	Files       []fileDict "files"
}

type fileDict struct {
	Length int64    "length"
	Md5Sum string   "md5sum"
	Path   []string "path"
}

func parse(filePath string) (*metaInfo, error) {
	m := new(metaInfo)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	obj, err := bencode.Decode(f)
	j, err := json.Marshal(obj)
	if err != nil {
		return nil, MalformedTorrentError
	}

	// Calculate sha1 hash of the info map.
	top, ok := obj.(map[string]interface{})
	if !ok {
		return nil, MalformedTorrentError
	}
	info, ok := top["info"]
	if !ok {
		return nil, MalformedTorrentError
	}
	err = json.Unmarshal(j, m)
	if err != nil {
		return nil, MalformedTorrentError
	}
	var b bytes.Buffer
	bencode.Marshal(&b, info)

	// Generate the info hash.
	hash := sha1.New()
	hash.Write(b.Bytes())
	m.InfoHash = string(hash.Sum(nil))
	return m, nil
}
