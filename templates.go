package main

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
	"github.com/julienschmidt/httprouter"
)

// DataFrames represents the data for the /frames template
type DataFrames struct {
	Paths []DataFramesPath
}

// DataFramesPath represents the path data for the /frames template
type DataFramesPath struct {
	Path    string
	Streams []DataFramesStream
}

// DataFrames represents the stream data for the /frames template
type DataFramesStream struct {
	Datasets []DataFramesDataset
	Name     string
}

// DataFrames represents the dataset data for the /frames template
type DataFramesDataset struct {
	BorderColor string
	Data        []DataFramesData
	Label       string
}

// DataFrames represents the data data for the /frames template
type DataFramesData struct {
	X float64
	Y float64
}

// templateData returns the data needed by the template
func templateData(name string, r *http.Request, p httprouter.Params) (d interface{}, err error) {
	switch name {
	case "/frames.html":
		// TODO Move
		var colors = map[string]string{
			"I": "rgba(255,99,132,1)",
			"P": "rgba(54, 162, 235, 1)",
			"B": "rgba(255, 206, 86, 1)",
		}

		// Loop in paths
		var o = DataFrames{}
		for _, p := range strings.Split(r.URL.Query().Get("paths"), ",") {
			// Init
			var dfp = DataFramesPath{Path: p}

			// Get streams
			var ss []astiffprobe.Stream
			astilog.Debugf("Getting streams of %s", p)
			if ss, err = ffprobe.Streams(context.Background(), p); err != nil {
				return
			}

			// Loop through streams
			for _, s := range ss {
				// Only analyze video
				if s.CodecType == astiffprobe.CodecTypeVideo {
					// Init
					var dfs = DataFramesStream{}

					// Get name
					dfs.Name = filepath.Base(p) + " - "
					if s.Bitrate > 0 {
						dfs.Name += strconv.Itoa(int(s.Bitrate/1024)) + "kb - "
					}
					dfs.Name += fmt.Sprintf("%dx%d", s.Width, s.Height)

					// Get frames
					var fs []astiffprobe.Frame
					astilog.Debugf("Getting frames of stream %d of %s", s.Index, p)
					if fs, err = ffprobe.Frames(context.Background(), p, s.Index); err != nil {
						return
					}

					// Loop through frames
					var dtts = make(map[string]*DataFramesDataset)
					for _, f := range fs {
						if _, ok := dtts[f.PictType]; !ok {
							dtts[f.PictType] = &DataFramesDataset{BorderColor: colors[f.PictType], Label: f.PictType}
						}
						var d = dtts[f.PictType]
						d.Data = append(d.Data, DataFramesData{X: f.BestEffortTimestampTime.Seconds(), Y: float64(f.PktSize) / f.PktDurationTime.Seconds() / 1024})
					}

					// Loop through datasets
					for _, d := range dtts {
						dfs.Datasets = append(dfs.Datasets, *d)
					}

					// Add stream data
					dfp.Streams = append(dfp.Streams, dfs)
				}
			}

			// Add path data
			o.Paths = append(o.Paths, dfp)
		}
		d = o
	}
	return
}
