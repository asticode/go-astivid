package astiffprobe

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Output represents the object FFProbe outputs
// https://ffmpeg.org/doxygen/2.7/structAVFrame.html
type Output struct {
	Frames  []Frame  `json:"frames"`
	Streams []Stream `json:"streams"`
}

// Bool represents a boolean in an int format
type Bool bool

// UnmarshalText implements the JSONUnmarshaler interface
// We need to use UnmarshalJSON instead of UnmarshalText otherwise it fails
func (bl *Bool) UnmarshalJSON(b []byte) (err error) {
	if string(b) == "1" {
		*bl = Bool(true)
		return
	}
	*bl = Bool(false)
	return
}

// Division represents a float in a division string format ("25/1")
type Division float64

// UnmarshalText implements the TextUnmarshaler interface
func (d *Division) UnmarshalText(b []byte) (err error) {
	var p = strings.Split(string(b), "/")
	var o float64
	if len(p) == 0 {
		err = fmt.Errorf("Invalid number of args for framerate %s", b)
		return
	} else if len(p) == 1 {
		if o, err = strconv.ParseFloat(p[0], 64); err == nil {
			*d = Division(o)
		}
	} else {
		var i1, i2 = float64(0), float64(0)
		if i1, err = strconv.ParseFloat(p[0], 64); err != nil {
			return
		}
		if i2, err = strconv.ParseFloat(p[1], 64); err != nil {
			return
		}
		if i1 == 0 || i2 == 0 {
			*d = 0
		} else {
			*d = Division(i1 / i2)
		}
	}
	return
}

// Duration represents a duration in a string format "1.203" such as the duration is 1.203s
type Duration struct {
	time.Duration
}

// UnmarshalText implements the TextUnmarshaler interface
func (d *Duration) UnmarshalText(b []byte) (err error) {
	var f float64
	if f, err = strconv.ParseFloat(string(b), 64); err != nil {
		return
	}
	*d = Duration{(time.Duration(f * 1e9))}
	return
}

// Hexadecimal represents an int in hexadecimal format
type Hexadecimal string

// Hexadecimal implements the TextUnmarshaler interface
func (h *Hexadecimal) UnmarshalText(b []byte) (err error) {
	var n int64
	if n, err = strconv.ParseInt(string(b), 0, 64); err != nil {
		return
	}
	*h = Hexadecimal(strconv.Itoa(int(n)))
	return
}

// Ratio represents a ration in the format "16:9"
type Ratio struct {
	Height, Width int
}

// UnmarshalText implements the TextUnmarshaler interface
func (r *Ratio) UnmarshalText(b []byte) (err error) {
	var items = strings.Split(string(b), ":")
	if len(items) < 2 {
		return fmt.Errorf("Invalid ration %s", b)
	}
	if r.Width, err = strconv.Atoi(items[0]); err != nil {
		return
	}
	if r.Height, err = strconv.Atoi(items[1]); err != nil {
		return
	}
	return
}
