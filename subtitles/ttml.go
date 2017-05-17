package astisub

import (
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Vars
var (
	regexpTTMLDurationFrames = regexp.MustCompile("\\:[\\d]+$")
)

// TTML represents a TTML
// https://www.speechpad.com/captions/ttml
type TTML struct {
	Framerate int            `xml:"frameRate,attr,omitempty"`
	Lang      string         `xml:"lang,attr,omitempty"`
	Regions   []TTMLRegion   `xml:"head>layout>region,omitempty"`
	Styles    []TTMLStyle    `xml:"head>styling>style,omitempty"`
	Subtitles []TTMLSubtitle `xml:"body>div>p"`
	XMLName   xml.Name       `xml:"tt"`
}

// TTMLHeader represents a TTML header
type TTMLHeader struct {
	ID     string     `xml:"id,attr,omitempty"`
	Style  string     `xml:"style,attr,omitempty"`
	Styles []xml.Attr `xml:",attr,omitempty"`
}

// UnmarshalXML implements the XML unmarshaler interface
func (h *TTMLHeader) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for _, a := range start.Attr {
		switch strings.ToLower(a.Name.Local) {
		case "id":
			h.ID = a.Value
		case "style":
			h.Style = a.Value
		default:
			h.Styles = append(h.Styles, a)
		}
	}
	return d.Skip()
}

// TTMLRegion represents a TTML region
type TTMLRegion struct {
	TTMLHeader
	XMLName xml.Name `xml:"region"`
}

// TTMLStyle represents a TTML style
type TTMLStyle struct {
	TTMLHeader
	XMLName xml.Name `xml:"style"`
}

// TTMLSubtitle represents a TTML subtitle
type TTMLSubtitle struct {
	Begin  *TTMLDuration `xml:"begin,attr"`
	End    *TTMLDuration `xml:"end,attr"`
	ID     string        `xml:"id,attr,omitempty"`
	Region string        `xml:"region,attr,omitempty"`
	Style  string        `xml:"style,attr,omitempty"`
	Styles []xml.Attr    `xml:",attr,omitempty"`
	Text   []TTMLText
}

// UnmarshalXML implements the XML unmarshaler interface
func (s *TTMLSubtitle) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	// Get attributes
	for _, a := range start.Attr {
		switch strings.ToLower(a.Name.Local) {
		case "begin":
			if s.Begin, err = newTTLMDuration(a.Value); err != nil {
				return
			}
		case "end":
			if s.End, err = newTTLMDuration(a.Value); err != nil {
				return
			}
		case "id":
			s.ID = a.Value
		case "region":
			s.Region = a.Value
		case "style":
			s.Style = a.Value
		default:
			s.Styles = append(s.Styles, a)
		}
	}

	// Get next tokens
	var t xml.Token
	for {
		// Get next token
		if t, err = d.Token(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		// Start element
		if se, ok := t.(xml.StartElement); ok {
			var e = TTMLText{}
			if err = d.DecodeElement(&e, &se); err != nil {
				return
			}
			s.Text = append(s.Text, e)
		} else if b, ok := t.(xml.CharData); ok {
			var str = strings.TrimSpace(string(b))
			if len(str) > 0 {
				s.Text = append(s.Text, TTMLText{Sentence: str})
			}
		}
	}
	return nil
}

// TTMLText represents a TTML text
// TODO Add MarshalXML
type TTMLText struct {
	Sentence string     `xml:",chardata"`
	Style    string     `xml:"style,attr,omitempty"`
	Styles   []xml.Attr `xml:",attr,omitempty"`
	XMLName  xml.Name
}

// UnmarshalXML implements the XML unmarshaler interface
func (t *TTMLText) UnmarshalXML(d *xml.Decoder, start xml.StartElement) (err error) {
	// Get XML name
	t.XMLName = start.Name

	// Get attributes
	for _, a := range start.Attr {
		switch strings.ToLower(a.Name.Local) {
		case "style":
			t.Style = a.Value
		default:
			t.Styles = append(t.Styles, a)
		}
	}

	// Get next tokens
	var tkn xml.Token
	for {
		// Get next token
		if tkn, err = d.Token(); err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		// Char data
		if b, ok := tkn.(xml.CharData); ok {
			var str = strings.TrimSpace(string(b))
			if len(str) > 0 {
				t.Sentence = str
			}
		}
	}
	return nil
}

// TTMLDuration represents a TTML duration
type TTMLDuration struct {
	d                 time.Duration
	frames, framerate int // Framerate is in frame/s
}

// newTTLMDuration creates a new TTMLDuration
// Possible formats are:
// - hh:mm:ss.mmm
// - hh:mm:ss:fff (fff being frames)
func newTTLMDuration(text string) (d *TTMLDuration, err error) {
	// hh:mm:ss:fff format
	d = &TTMLDuration{}
	if indexes := regexpTTMLDurationFrames.FindStringIndex(text); indexes != nil {
		// Parse frames
		var s = text[indexes[0]+1 : indexes[1]]
		if d.frames, err = strconv.Atoi(s); err != nil {
			err = errors.Wrapf(err, "atoi %s failed", s)
			return
		}

		// Update text
		text = text[:indexes[0]] + ".000"
	}
	d.d, err = parseDuration(text, ".")
	return
}

// Duration returns the TTML Duration's time.Duration
func (d TTMLDuration) Duration() time.Duration {
	if d.framerate > 0 {
		return d.d + time.Duration(float64(d.frames)/float64(d.framerate)*1e9)*time.Nanosecond
	}
	return d.d
}

// MarshalText implements the TextMarshaler interface
func (t *TTMLDuration) MarshalText() ([]byte, error) {
	return []byte(formatDuration(t.Duration(), ".")), nil
}

// ReadFromTTML parses a .ttml content
// TODO Add region and style to subtitle as well
func ReadFromTTML(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = &Subtitles{}

	// Unmarshal XML
	var ttml TTML
	if err = xml.NewDecoder(i).Decode(&ttml); err != nil {
		return
	}

	// Loop through subtitles
	for _, s := range ttml.Subtitles {
		// TODO Get line instead
		// Get text
		var text []string
		for _, t := range s.Text {
			// TODO Is is the correct way to handle br ?
			if strings.ToLower(t.XMLName.Local) != "br" {
				text = append(text, t.Sentence)
			}
		}

		// Update framerate
		s.Begin.framerate = ttml.Framerate
		s.End.framerate = ttml.Framerate

		// Append subtitle
		o.Items = append(o.Items, &Subtitle{
			EndAt:   s.End.Duration(),
			StartAt: s.Begin.Duration(),
		})
	}
	return
}

// WriteToTTML writes subtitles in .ttml format
func (s Subtitles) WriteToTTML(o io.Writer) (err error) {
	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		return ErrNoSubtitlesToWrite
	}

	// Loop through items
	var ttml = TTML{}
	for _, sub := range s.Items {
		// TODO Init TTML text
		/*
			var text = []TTMLText{}
			for _, t := range sub.Text {
				text = append(text, TTMLText{Sentence: t})
			}
		*/

		// Append subtitle
		// TODO Add text
		ttml.Subtitles = append(ttml.Subtitles, TTMLSubtitle{
			Begin: &TTMLDuration{d: sub.StartAt},
			End:   &TTMLDuration{d: sub.EndAt},
			// Text:  text,
		})
	}

	// Marshal XML
	var e = xml.NewEncoder(o)
	e.Indent("", "    ")
	return e.Encode(ttml)
}
