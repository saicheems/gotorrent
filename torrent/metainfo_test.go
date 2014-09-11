package torrent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := map[string]*MetaInfo{
		"d8:announce14:http://sai.com13:creation datei1234e7:comment2:hi10:created by3:sai8:encoding3:idk4:infod12:piece lengthi4eee": &MetaInfo{Announce: "http://sai.com", CreationDate: 1234, Comment: "hi", CreatedBy: "sai", Encoding: "idk", Info: InfoDict{PieceLength: 4}, InfoHash: "\xa4t\xe1z[)f\x11\xe2c\x02E\vג\n\xa5\x81\xbf\xa5"},
	}

	for key, val := range tests {
		m, err := Parse(strings.NewReader(key))

		assert := assert.New(t)
		assert.Nil(err)
		assert.Equal(m, val)
	}
}

func TestParseError(t *testing.T) {
	test := "d8:announce14:http://sai.com13:creation date4:12347:comment2:hi10:created by3:sai8:encoding3:idke"

	_, err := Parse(strings.NewReader(test))

	assert := assert.New(t)
	assert.NotNil(err)
}
