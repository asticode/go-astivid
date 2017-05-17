package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"encoding/xml"

	"github.com/asticode/go-astivid/subtitles"
	"github.com/stretchr/testify/assert"
)

func TestTTMLHeader(t *testing.T) {
	type Test struct {
		astisub.TTMLHeader
		XMLName xml.Name `xml:"test"`
	}
	var r = Test{}
	err := xml.Unmarshal([]byte("<test id=\"id\" style=\"style\" key=\"value\"></test>"), &r)
	assert.NoError(t, err)
	assert.Equal(t, Test{TTMLHeader: astisub.TTMLHeader{ID: "id", Style: "style", Styles: []xml.Attr{{Name: xml.Name{Local: "key"}, Value: "value"}}}}, r)
	b, err := xml.Marshal(r)
	assert.NoError(t, err)
	assert.Equal(t, "<test id=\"id\" style=\"style\" key=\"value\"></test>", string(b))
}

func TestTTMLSubtitle(t *testing.T) {
	var s = astisub.TTMLSubtitle{}
	err := xml.Unmarshal([]byte(`<p begin="00:00:01.20" end="00:00:07.84" id="id" tts:textAlign="center" style="style_1" region="region_1">Twinkle, twinkle, little bat!<br/>How <span tts:fontStyle="italic">I wonder</span> where you're at!</p>`), &s)
	assert.NoError(t, err)
	assert.Equal(t, "", s)
	b, err := xml.Marshal(s)
	assert.NoError(t, err)
	assert.Equal(t, "", string(b))
}

func TestTTML(t *testing.T) {
	// Open
	s, err := astisub.Open("./testdata/example-in.ttml")
	assert.NoError(t, err)
	assertSubtitles(t, s)

	// No subtitles to write
	w := &bytes.Buffer{}
	err = astisub.Subtitles{}.WriteToTTML(w)
	assert.EqualError(t, err, astisub.ErrNoSubtitlesToWrite.Error())

	// Write
	c, err := ioutil.ReadFile("./testdata/example-out.ttml")
	assert.NoError(t, err)
	err = s.WriteToTTML(w)
	assert.NoError(t, err)
	assert.Equal(t, string(c), w.String())
	assert.False(t, true)
}
