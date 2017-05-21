package astisub

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"
)

// Constants
const (
	timeBoundariesSeparator = " --> "
)

// Vars
var (
	bytesTimeBoundariesSeparator = []byte(timeBoundariesSeparator)
)

// parseDurationSRT parses an .srt duration
func parseDurationSRT(i string) (time.Duration, error) {
	return parseDuration(i, ",")
}

// ReadFromSRT parses an .srt content
func ReadFromSRT(i io.Reader) (o *Subtitles, err error) {
	// Init
	o = &Subtitles{}
	var scanner = bufio.NewScanner(i)

	// Scan
	var line string
	var s = &Subtitle{}
	for scanner.Scan() {
		// Fetch line
		line = scanner.Text()

		// Line contains time boundaries
		if strings.Contains(line, timeBoundariesSeparator) {
			// Remove last item of previous subtitle since it's the index
			s.Lines = s.Lines[:len(s.Lines)-1]

			// Remove trailing empty lines
			if len(s.Lines) > 0 {
				for i := len(s.Lines) - 1; i >= 0; i-- {
					if len(s.Lines[i]) > 0 {
						for j := len(s.Lines[i]) - 1; j >= 0; j-- {
							if len(s.Lines[i][j].Sentence) == 0 {
								s.Lines[i] = s.Lines[i][:j]
							} else {
								break
							}
						}
						if len(s.Lines[i]) == 0 {
							s.Lines = s.Lines[:i]
						}

					}
				}
			}

			// Init subtitle
			s = &Subtitle{}

			// Fetch time boundaries
			boundaries := strings.Split(line, timeBoundariesSeparator)
			if s.StartAt, err = parseDurationSRT(boundaries[0]); err != nil {
				return
			}
			if s.EndAt, err = parseDurationSRT(boundaries[1]); err != nil {
				return
			}

			// Append subtitle
			o.Items = append(o.Items, s)
		} else {
			// Add text
			s.Lines = append(s.Lines, []Text{{Sentence: line}})
		}
	}
	return
}

// formatDurationSRT formats an .srt duration
func formatDurationSRT(i time.Duration) string {
	return formatDuration(i, ",")
}

// WriteToSRT writes subtitles in .srt format
func (s Subtitles) WriteToSRT(o io.Writer) (err error) {
	// Init
	var c []byte

	// Do not write anything if no subtitles
	if len(s.Items) == 0 {
		err = ErrNoSubtitlesToWrite
		return
	}

	// Add BOM header
	c = append(c, bytesBOM...)

	// Loop through subtitles
	for k, v := range s.Items {
		// Init content
		c = append(c, []byte(strconv.Itoa(k+1))...)
		c = append(c, bytesLineSeparator...)
		c = append(c, []byte(formatDurationSRT(v.StartAt))...)
		c = append(c, bytesTimeBoundariesSeparator...)
		c = append(c, []byte(formatDurationSRT(v.EndAt))...)
		c = append(c, bytesLineSeparator...)

		// Loop through lines
		for _, l := range v.Lines {
			// Loop through texts
			var ts [][]byte
			for _, t := range l {
				ts = append(ts, []byte(t.Sentence))
			}
			c = append(c, bytes.Join(ts, bytesSpace)...)
			c = append(c, bytesLineSeparator...)
		}

		// Add new line
		c = append(c, bytesLineSeparator...)
	}

	// Remove last new line
	c = c[:len(c)-1]

	// Write
	if _, err = o.Write(c); err != nil {
		return
	}
	return
}
