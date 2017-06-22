package astisub_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"encoding/xml"

	"time"

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
	assert.Equal(t, time.Second+200*time.Millisecond, s.Begin.Duration())
	assert.Equal(t, 7*time.Second+840*time.Millisecond, s.End.Duration())
	assert.Equal(t, "id", s.ID)
	assert.Equal(t, "region_1", s.Region)
	assert.Equal(t, "style_1", s.Style)
	assert.Equal(t, []xml.Attr{{Name: xml.Name{Space: "tts", Local: "textAlign"}, Value: "center"}}, s.Styles)
	assert.Equal(t, []astisub.TTMLText{{Sentence: "Twinkle, twinkle, little bat!"}, {XMLName: xml.Name{Local: "br"}}, {Sentence: "How"}, {Sentence: "I wonder", Styles: []xml.Attr{{Name: xml.Name{Space: "tts", Local: "fontStyle"}, Value: "italic"}}, XMLName: xml.Name{Local: "span"}}, {Sentence: "where you're at!"}}, s.Text)
}

func TestTTML(t *testing.T) {
	// Open
	s, err := astisub.Open("./testdata/example-in.ttml")
	assert.NoError(t, err)
	assertSubtitleItems(t, s)

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
