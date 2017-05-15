package main

import (
	"flag"

	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron/loader"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
	"github.com/pkg/errors"
)

// Vars
var (
	ffprobe *astiffprobe.FFProbe
)

func main() {
	// Init
	flag.Parse()
	var c = newConfiguration()
	astilog.SetLogger(astilog.New(c.Logger))
	ffprobe = astiffprobe.New(c.FFProbe)

	// Provision
	var err error
	if err = provision(); err != nil {
		astilog.Fatal(errors.Wrap(err, "provisioning failed"))
	}

	// Serve
	var ln = serve(c)

	// Create astilectron
	var a *astilectron.Astilectron
	if a, err = astilectron.New(astilectron.Options{AppName: "Astivid"}); err != nil {
		astilog.Fatal(errors.Wrap(err, "creating new astilectron failed"))
	}
	defer a.Close()
	a.HandleSignals()

	// Start loader
	var l = astiloader.NewForAstilectron(a)
	go l.Start()

	// Start
	if err = a.Start(); err != nil {
		astilog.Fatal(errors.Wrap(err, "starting astilectron failed"))
	}

	// Create window
	var w *astilectron.Window
	if w, err = a.NewWindow("http://"+ln.Addr().String()+"/templates/index", &astilectron.WindowOptions{BackgroundColor: astilectron.PtrStr("#333"), Center: astilectron.PtrBool(true), Height: astilectron.PtrInt(600), Width: astilectron.PtrInt(600)}); err != nil {
		astilog.Fatal(errors.Wrap(err, "new window failed"))
	}
	if err = w.Create(); err != nil {
		astilog.Fatal(errors.Wrap(err, "creating window failed"))
	}

	// Blocking pattern
	a.Wait()
}
