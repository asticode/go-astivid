package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/asticode/go-astivid/subtitles"
	"github.com/stretchr/testify/assert"
)

func TestSRT(t *testing.T) {
	// Init
	const path = "./testdata/example.srt"

	// Open
	s, err := astisub.Open(path)
	assert.NoError(t, err)
	assertSubtitles(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToSRT(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile(path)
	assert.NoError(t, err)
	err = s.WriteToSRT(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
}
