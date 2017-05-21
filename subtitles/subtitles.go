package astisub

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Bytes
var (
	bytesBOM           = []byte{239, 187, 191}
	bytesColon         = []byte(":")
	bytesComma         = []byte(",")
	bytesLineSeparator = []byte("\n")
	bytesPeriod        = []byte(".")
)

// Errors
var (
	ErrInvalidExtension   = errors.New("Invalid extension")
	ErrNoSubtitlesToWrite = errors.New("No subtitles to write")
)

// Open opens a subtitle file
func Open(src string) (s *Subtitles, err error) {
	// Open the file
	var f *os.File
	if f, err = os.Open(src); err != nil {
		err = errors.Wrapf(err, "opening %s failed", src)
		return
	}
	defer f.Close()

	// Parse the content
	switch filepath.Ext(src) {
	case ".srt":
		s, err = ReadFromSRT(f)
	case ".ttml":
		//s, err = ReadFromTTML(f)
	case ".vtt":
		//s, err = ReadFromVTT(f)
	default:
		err = ErrInvalidExtension
	}
	return
}

// Subtitles represents an ordered list of subtitles with formatting
type Subtitles struct {
	Items   []*Subtitle
	Regions bool
	Styles  bool
}

// Subtitle represents a text to show between 2 time boundaries
type Subtitle struct {
	EndAt   time.Duration
	StartAt time.Duration
	Text    []string
}

// Add adds a duration to each time boundaries. As in the time package, duration can be negative.
func (s *Subtitles) Add(d time.Duration) {
	for _, v := range s.Items {
		v.EndAt += d
		v.StartAt += d
	}
}

// Duration returns the subtitles duration
func (s Subtitles) Duration() time.Duration {
	if len(s.Items) == 0 {
		return time.Duration(0)
	}
	return s.Items[len(s.Items)-1].EndAt
}

// ForceDuration updates the subtitles duration.
// If requested duration is bigger, then we create a dummy item.
// If requested duration is smaller, then we remove useless items and we cut the last item or add a dummy item.
func (s *Subtitles) ForceDuration(d time.Duration) {
	// Requested duration is the same as the subtitles'one
	if s.Duration() == d {
		return
	}

	// Requested duration is bigger than subtitles'one
	if s.Duration() > d {
		// Find last item before input duration and update end at
		var lastIndex = -1
		for index, i := range s.Items {
			// Start at is bigger than input duration, we've found the last item
			if i.StartAt >= d {
				lastIndex = index
				break
			} else if i.EndAt > d {
				s.Items[index].EndAt = d
			}
		}

		// Last index has been found
		if lastIndex != -1 {
			s.Items = s.Items[:lastIndex]
		}
	}

	// Add dummy item
	if s.Duration() < d {
		s.Items = append(s.Items, &Subtitle{EndAt: d, StartAt: d, Text: []string{"..."}})
	}
}

// Fragment fragments subtitles with a specific fragment duration
func (s *Subtitles) Fragment(f time.Duration) {
	// Nothing to fragment
	if len(s.Items) == 0 {
		return
	}

	// Here we want to simulate fragments of duration f until there are no subtitles left in that period of time
	var fragmentStartAt, fragmentEndAt = time.Duration(0), f
	for fragmentStartAt < s.Items[len(s.Items)-1].EndAt {
		// We loop through subtitles and process the ones that either contain the fragment start at,
		// or contain the fragment end at
		//
		// It's useless processing subtitles contained between fragment start at and end at
		//             |____________________|             <- subtitle
		//           |                        |
		//   fragment start at        fragment end at
		for i, sub := range s.Items {
			// Init
			var newSub = &Subtitle{}
			*newSub = *sub

			// A switch is more readable here
			switch {
			// Subtitle contains fragment start at
			// |____________________|                         <- subtitle
			//           |                        |
			//   fragment start at        fragment end at
			case sub.StartAt < fragmentStartAt && sub.EndAt > fragmentStartAt:
				sub.StartAt = fragmentStartAt
				newSub.EndAt = fragmentStartAt
			// Subtitle contains fragment end at
			//                         |____________________| <- subtitle
			//           |                        |
			//   fragment start at        fragment end at
			case sub.StartAt < fragmentEndAt && sub.EndAt > fragmentEndAt:
				sub.StartAt = fragmentEndAt
				newSub.EndAt = fragmentEndAt
			default:
				continue
			}

			// Insert new sub
			s.Items = append(s.Items[:i], append([]*Subtitle{newSub}, s.Items[i:]...)...)
		}

		// Update fragments boundaries
		fragmentStartAt += f
		fragmentEndAt += f
	}
}

// IsEmpty returns whether the subtitles are empty
func (s Subtitles) IsEmpty() bool {
	return len(s.Items) == 0
}

// Merge merges subtitles i into subtitles s
func (s *Subtitles) Merge(i Subtitles) {
	// Loop through input subtitles
	for _, subInput := range i.Items {
		var lastIndex int
		var inserted bool
		// Loop through parent subtitles
		for index, subParent := range s.Items {
			// Input sub is after parent sub
			if subInput.StartAt < subParent.StartAt {
				s.Items = append(s.Items[:lastIndex+1], append([]*Subtitle{subInput}, s.Items[lastIndex+1:]...)...)
				inserted = true
				break
			}
			lastIndex = index
		}
		if !inserted {
			s.Items = append(s.Items, subInput)
		}
	}
}

// Write writes subtitles to a file
func (s Subtitles) Write(dst string) (err error) {
	// Create the file
	var f *os.File
	if f, err = os.Create(dst); err != nil {
		return
	}
	defer f.Close()

	// Write the content
	switch filepath.Ext(dst) {
	case ".srt":
		err = s.WriteToSRT(f)
	case ".tml":
		//err = s.WriteToTTML(f)
	default:
		err = ErrInvalidExtension
	}
	return
}

// parseDuration parses a duration in "00:00:00.000" or "00:00:00,000" format
func parseDuration(i, millisecondSep string) (o time.Duration, err error) {
	// Split milliseconds
	var parts = strings.Split(i, millisecondSep)
	if len(parts) != 2 {
		err = fmt.Errorf("No milliseconds detected in %s", i)
		return
	}
	if len(parts[1]) != 3 {
		err = fmt.Errorf("Invalid number of millisecond digits detected in %s", i)
		return
	}

	// Parse milliseconds
	var milliseconds int
	var s = strings.TrimSpace(parts[1])
	if milliseconds, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Split hours, minutes and seconds
	parts = strings.Split(strings.TrimSpace(parts[0]), ":")
	if len(parts) != 3 {
		err = fmt.Errorf("No hours, minutes or seconds detected in %s", i)
		return
	}

	// Parse seconds
	var seconds int
	s = strings.TrimSpace(parts[2])
	if seconds, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Parse minutes
	var minutes int
	s = strings.TrimSpace(parts[1])
	if minutes, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Parse hours
	var hours int
	s = strings.TrimSpace(parts[0])
	if hours, err = strconv.Atoi(s); err != nil {
		err = errors.Wrapf(err, "atoi of %s failed", s)
		return
	}

	// Generate output
	o = time.Duration(milliseconds)*time.Millisecond + time.Duration(seconds)*time.Second + time.Duration(minutes)*time.Minute + time.Duration(hours)*time.Hour
	return
}

// formatDurationSRT formats a duration
func formatDuration(i time.Duration, millisecondSep string) (s string) {
	// Parse hours
	var hours = int(i / time.Hour)
	var n = i % time.Hour
	if hours < 10 {
		s += "0"
	}
	s += strconv.Itoa(hours) + ":"

	// Parse minutes
	var minutes = int(n / time.Minute)
	n = i % time.Minute
	if minutes < 10 {
		s += "0"
	}
	s += strconv.Itoa(minutes) + ":"

	// Parse seconds
	var seconds = int(n / time.Second)
	n = i % time.Second
	if seconds < 10 {
		s += "0"
	}
	s += strconv.Itoa(seconds) + millisecondSep

	// Parse milliseconds
	var milliseconds = int(n / time.Millisecond)
	if milliseconds < 10 {
		s += "00"
	} else if milliseconds < 100 {
		s += "0"
	}
	s += strconv.Itoa(milliseconds)
	return
}
