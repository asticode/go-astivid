package main

import (
	"context"
	"flag"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/flag"
	"github.com/asticode/go-astivid/ffprobe"
)

// Flags
var (
	inputs = &astiflag.Strings{}
)

func main() {
	// Parse flags
	flag.Var(inputs, "i", "the input file(s)")
	flag.Parse()

	// TODO Provision ffprobe + ffmpeg =>
	// - Linux: https://www.johnvansickle.com/ffmpeg/
	// - Mac: https://evermeet.cx/ffmpeg/
	// - Windows: https://ffmpeg.zeranoe.com/builds/

	// Init
	var c = NewConfiguration()
	astilog.SetLogger(astilog.New(c.Logger))
	var ffprobe = astiffprobe.New(c.FFProbe)

	// Get frames
	var f []astiffprobe.Frame
	var err error
	if f, err = ffprobe.Frames(context.Background(), (*inputs)[0], 0); err != nil {
		astilog.Fatalf("Getting frames of %s failed: %s", (*inputs)[0], err)
	}
	astilog.Debugf("Frames are %+v", f[:3])
}
