package main

import (
	"flag"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron/bootstrap"
	"github.com/asticode/go-astilectron/loader"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
)

// Vars
var (
	ffprobe *astiffprobe.FFProbe
)

//go:generate go-bindata -pkg $GOPACKAGE -o resources.go resources/...
func main() {
	// Init
	flag.Parse()
	var c = newConfiguration()
	astilog.SetLogger(astilog.New(c.Logger))
	ffprobe = astiffprobe.New(c.FFProbe)

	// Run bootstrap
	if err := bootstrap.Run(bootstrap.Options{
		AdaptRouter: adaptRouter,
		AstilectronOptions: astilectron.Options{
			AppName: "Astivid",
		},
		CustomProvision: provision,
		Homepage:        "/templates/index",
		RestoreAssets:   RestoreAssets,
		StartLoader: func(a *astilectron.Astilectron) {
			var l = astiloader.NewForAstilectron(a)
			go l.Start()
		},
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
