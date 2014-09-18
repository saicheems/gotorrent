package torrent

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"

	"github.com/jackpal/bencode-go"
)

// MalformedTorrentError is the error returned when a torrent file couldn't be parsed.
var MalformedTorrentError = fmt.Errorf("Malformed torrent file.")

// MetaInfo implements the contents and file structure of a .torrent file.
type MetaInfo struct {
	Info         InfoDict   "info"
	Announce     string     "announce"
	AnnounceList [][]string "announce-list"
	CreationDate int64      "creation date"
	Comment      string     "comment"
	CreatedBy    string     "created by"
	Encoding     string     "encoding"

	InfoHash string
}

// InfoDict implements the info dictionary portion of a .torrent metainfo.
type InfoDict struct {
	PieceLength int64      "piece length"
	Pieces      string     "pieces"
	Private     string     "private"
	Name        string     "name"
	Length      int64      "length"
	Md5Sum      string     "md5sum"
	Files       []FileDict "files"
}

// FileDict implements the information encoded in a multi-file torrent.
type FileDict struct {
	Length int64    "length"
	Md5Sum string   "md5sum"
	Path   []string "path"
}

// Parse returns a MetaInfo struct filled in with data from the input stream. The input stream
// should be a bencoded torrent file. An error is raised if there is a problem in parsing.
func Parse(r io.Reader) (*MetaInfo, error) {
	m := new(MetaInfo)
	// TODO: This will backfire if the torrent file is for some reason too large to fit in
	// memory.
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	err := bencode.Unmarshal(bytes.NewReader(buf.Bytes()), m)
	obj, err := bencode.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, MalformedTorrentError
	}
	m.InfoHash, err = computeSha1Hash(obj)
	if err != nil {
		return nil, err
	}
	// TODO: Support this!
	if m.Info.Files != nil {
		return nil, errors.New("no support for multi-file torrents yet :(")
	}
	return m, nil
}

// computeSha1Hash finds the info dict in bencoded dictionary and returns its sha1 hash.
func computeSha1Hash(obj interface{}) (string, error) {
	// Calculate sha1 hash of the info map.
	top, ok := obj.(map[string]interface{})
	if !ok {
		return "", MalformedTorrentError
	}
	info, ok := top["info"]
	if !ok {
		return "", MalformedTorrentError
	}
	var b bytes.Buffer
	bencode.Marshal(&b, info)
	// Generate the info hash.
	hash := sha1.New()
	hash.Write(b.Bytes())
	return string(hash.Sum(nil)), nil
}
