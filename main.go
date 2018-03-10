package main

import (
	"flag"

	"github.com/asticode/go-astiffmpeg"
	"github.com/asticode/go-astiffprobe"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
)

// Vars
var (
	AppName string
	BuiltAt string
	c       Configuration
	debug   = flag.Bool("d", false, "if yes, the app is in debug mode")
	ffmpeg  *astiffmpeg.FFMpeg
	ffprobe *astiffprobe.FFProbe
)

// TODO Add subtitle actions => add / convert
func main() {
	// Init
	flag.Parse()
	c = newConfiguration()
	astilog.SetLogger(astilog.New(c.Logger))
	ffmpeg = astiffmpeg.New(c.FFMpeg)
	ffprobe = astiffprobe.New(c.FFProbe)

	// TODO Provision ffprobe + ffmpeg =>
	// - Linux: https://www.johnvansickle.com/ffmpeg/
	// - Mac: https://evermeet.cx/ffmpeg/
	// - Windows: https://ffmpeg.zeranoe.com/builds/

	// Run bootstrap
	if err := bootstrap.Run(bootstrap.Options{
		Asset: Asset,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/gopher.icns",
			AppIconDefaultPath: "resources/gopher.png",
		},
		Debug:          true,
		Homepage:       "index.html",
		MessageHandler: handleMessages,
		RestoreAssets:  RestoreAssets,
		WindowOptions: &astilectron.WindowOptions{
			BackgroundColor: astilectron.PtrStr("#333"),
			Center:          astilectron.PtrBool(true),
			Height:          astilectron.PtrInt(600),
			Width:           astilectron.PtrInt(600),
		},
	}); err != nil {
		astilog.Fatal(err)
	}
}
