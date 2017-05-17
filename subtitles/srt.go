package astisub

import (
	"bufio"
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
func parseDurationSRT(i string) (o time.Duration, err error) {
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
			s.Text = s.Text[:len(s.Text)-1]

			// Remove trailing empty lines
			if len(s.Text) > 0 {
				for i := len(s.Text) - 1; i > 0; i-- {
					if s.Text[i] == "" {
						s.Text = s.Text[:i]
					} else {
						break
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
			s.Text = append(s.Text, line)
		}
	}
	return
}

// formatDurationSRT formats an .srt duration
func formatDurationSRT(i time.Duration) (s string) {
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

		// Add text
		for _, t := range v.Text {
			c = append(c, []byte(t)...)
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
