package main

import (
	"context"
	"flag"
	"os/exec"

	"os"
	"path/filepath"

	"io/ioutil"
	"strings"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitesseract"
	"github.com/asticode/go-astitools/os"
	"github.com/pkg/errors"
)

// Flags
var (
	inputPath            = flag.String("i", "", "the input path")
	workingDirectoryPath = flag.String("w", "", "the working directory path")
)

func main() {
	// Init
	flag.Parse()
	astilog.FlagInit()

	// Validate input path
	if len(*inputPath) == 0 {
		astilog.Fatal("Use -i to indicate an input path")
	}

	// Create working directory
	var err error
	if len(*workingDirectoryPath) == 0 {
		if *workingDirectoryPath, err = astios.TempDir("astivid"); err != nil {
			astilog.Fatal(errors.Wrap(err, "creating working directory failed"))
		}
		defer os.RemoveAll(*workingDirectoryPath)
	}
	astilog.Debugf("Working directory is %s", *workingDirectoryPath)

	// Init tesseract
	var tst *astitesseract.Tesseract
	if tst, err = astitesseract.New(astitesseract.Options{Languages: []string{"spa"}}); err != nil {
		astilog.Fatal(errors.Wrap(err, "astitesseract.New failed"))
	}
	defer tst.Close()

	// Init cmd
	var cmd = exec.CommandContext(context.Background(), "ffmpeg", "-i", *inputPath, "-filter:v", "crop=iw:ih-250:0:250", "-r", "1/1", filepath.Join(*workingDirectoryPath, "frame%03d.jpg"))
	var b []byte
	if b, err = cmd.CombinedOutput(); err != nil {
		astilog.Fatalf("%s: %s", err, b)
	}

	// Read dir
	var fs []os.FileInfo
	if fs, err = ioutil.ReadDir(*workingDirectoryPath); err != nil {
		astilog.Fatal(errors.Wrapf(err, "reading dir %s failed", *workingDirectoryPath))
	}

	// Loop through files
	for _, f := range fs {
		// Only process .jpg files starting with "frame"
		if f.IsDir() || filepath.Ext(f.Name()) != ".jpg" || !strings.HasPrefix(f.Name(), "frame") {
			continue
		}

		// Get UTF8 text
		var p = filepath.Join(*workingDirectoryPath, f.Name())
		var txt = tst.GetUTF8Text(p)
		astilog.Debugf("%s: %s", p, txt)
	}
}
