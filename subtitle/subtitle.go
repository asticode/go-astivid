package astisub

import (
	"os"
	"time"

	"github.com/pkg/errors"
)

// Vars
var (
	BytesBOM              = []byte{239, 187, 191}
	bytesColon            = []byte(":")
	bytesComma            = []byte(",")
	bytesLineSeparator    = []byte("\n")
	bytesPeriod           = []byte(".")
	ErrInvalidExtension   = errors.New("Invalid extension")
	ErrNoSubtitlesToWrite = errors.New("No subtitles to write")
)

// Open opens a subtitle file
func Open(name string) (s *Subtitles, err error) {
	// Open the file
	var f *os.File
	if f, err = os.Open(name); err != nil {
		err = errors.Wrapf(err, "opening %s failed", name)
		return
	}
	defer f.Close()

	// Parse the content
	/*
		switch filepath.Ext(name) {
		case ".srt":
			s, err = ReadFromSRT(f)
		case ".ttml":
			s, err = ReadFromTTML(f)
		case ".vtt":
			s, err = ReadFromVTT(f)
		default:
			err = ErrInvalidExtension
		}
	*/
	return
}

// Subtitles represents an ordered list of subtitles with formatting
type Subtitles struct {
	Items   []*Subtitle
	Regions bool
	Styles  bool
}

// Duration returns the subtitles duration
func (s Subtitles) Duration() time.Duration {
	if len(s.Items) == 0 {
		return time.Duration(0)
	}
	return s.Items[len(s.Items)-1].EndAt
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

// Empty returns whether the subtitles are empty
func (s Subtitles) Empty() bool {
	return len(s.Items) == 0
}

// ForceDuration updates the subtitles duration
// If input duration is bigger, then we create a dummy item
// If input duration is smaller, then we remove useless items and we cut the last item
func (s *Subtitles) ForceDuration(d time.Duration) {
	// Input duration is the same as the subtitles'one
	if s.Duration() == d {
		return
	}

	// Input duration is bigger than subtitles'one
	if s.Duration() < d {
		// Add dummy item
		s.Items = append(s.Items, &Subtitle{EndAt: d, StartAt: d, Text: []string{"..."}})
	} else {
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
func (s Subtitles) Write(name string) (err error) {
	// Create the file
	var f *os.File
	if f, err = os.Create(name); err != nil {
		return
	}
	defer f.Close()

	// Write the content
	/*
		switch filepath.Ext(name) {
		case ".srt":
			err = s.WriterToSRT(f)
		case ".tml":
			err = s.WriteToTTML(f)
		default:
			err = ErrInvalidExtension
		}
	*/
	return
}
