package main

import (
	"os"
	"path/filepath"

	"github.com/asticode/go-astilog"
	"github.com/pkg/errors"
)

//go:generate go-bindata -pkg $GOPACKAGE -o resources.go resources/...
func provision() (err error) {
	// Get executable path
	var p string
	if p, err = os.Executable(); err != nil {
		return errors.Wrap(err, "getting executable path failed")
	}
	p = filepath.Dir(p)

	// Provision resources
	var pr = filepath.Join(p, "resources")
	if _, err = os.Stat(pr); os.IsNotExist(err) {
		// Restore assets
		astilog.Debugf("Restoring assets in %s", p)
		if err = RestoreAssets(p, "resources"); err != nil {
			return errors.Wrapf(err, "restoring assets in %s failed", p)
		}
	} else if err != nil {
		return errors.Wrapf(err, "stating %s failed", pr)
	}

	// TODO Provision ffprobe + ffmpeg =>
	// - Linux: https://www.johnvansickle.com/ffmpeg/
	// - Mac: https://evermeet.cx/ffmpeg/
	// - Windows: https://ffmpeg.zeranoe.com/builds/
	return
}
