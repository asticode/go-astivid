package main

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/asticode/go-astichartjs"
	"github.com/asticode/go-astiffprobe"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/ptr"
	"github.com/pkg/errors"
)

// handleMessages handles messages
func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (_ interface{}, _ error) {
	// Get payload
	switch m.Name {
	case "get.frames":
		return handleGetFrames(m)
	}
	return
}

// handleGetFrames handles the "get.frames" event
func handleGetFrames(i bootstrap.MessageIn) (payload interface{}, err error) {
	// Decode input
	var ps []string
	if err = json.Unmarshal(i.Payload, &ps); err != nil {
		err = errors.Wrap(err, "decoding input failed")
		return
	}

	// Loop through paths
	var cs = []astichartjs.Chart{}
	for _, p := range ps {
		// Retrieve streams
		var ss []astiffprobe.Stream
		astilog.Debugf("Retrieving streams of %s", p)
		if ss, err = ffprobe.Streams(context.Background(), p); err != nil {
			err = errors.Wrapf(err, "retrieving streams of %s failed", p)
			return
		}

		// Loop through streams
		for _, s := range ss {
			// Only analyze video
			if s.CodecType != astiffprobe.CodecTypeVideo {
				astilog.Debugf("Stream %d of %s is not a video, moving on...", s.Index, p)
				continue
			}

			// Init chart
			// TODO Increase title font size
			// TODO Resize doesn't work on MacOSX
			// TODO Disable border on btn click
			var c = astichartjs.Chart{
				Data: &astichartjs.Data{},
				Options: &astichartjs.Options{
					Scales: &astichartjs.Scales{
						XAxes: []astichartjs.Axis{
							{
								Position: astichartjs.ChartAxisPositionsBottom,
								ScaleLabel: &astichartjs.ScaleLabel{
									Display:     astiptr.Bool(true),
									LabelString: "Timestamp (s)",
								},
								Type: astichartjs.ChartAxisTypesLinear,
							},
						},
						YAxes: []astichartjs.Axis{
							{
								ScaleLabel: &astichartjs.ScaleLabel{
									Display:     astiptr.Bool(true),
									LabelString: "Bitrate (kb/s)",
								},
							},
						},
					},
					Title: &astichartjs.Title{Display: astiptr.Bool(true)},
				},
				Type: astichartjs.ChartTypeLine,
			}

			// Get chart title
			c.Options.Title.Text = filepath.Base(p) + " - "
			if s.Bitrate > 0 {
				c.Options.Title.Text += strconv.Itoa(int(s.Bitrate/1024)) + "kb - "
			}
			c.Options.Title.Text += fmt.Sprintf("%dx%d", s.Width, s.Height)

			// Retrieve frames
			var fs []astiffprobe.Frame
			astilog.Debugf("Retrieving frames of stream %d of %s", s.Index, p)
			if fs, err = ffprobe.Frames(context.Background(), p, s.Index); err != nil {
				err = errors.Wrapf(err, "retrieving frames of stream %d of %s", s.Index, p)
				return
			}

			// Loop through frames
			var ds = map[string]*astichartjs.Dataset{
				"avg": {
					BackgroundColor: astichartjs.ChartBackgroundColorGreen,
					BorderColor:     astichartjs.ChartBorderColorGreen,
					Label:           "Average per GOP",
				},
			}
			var vs []astichartjs.DataPoint
			for _, f := range fs {
				// Compute average
				if f.PictType == "I" && len(vs) > 0 {
					var sum float64
					for _, dp := range vs {
						sum += dp.Y
					}
					ds["avg"].Data = append(ds["avg"].Data, astichartjs.DataPoint{
						X: vs[0].X,
						Y: sum / float64(len(vs)),
					})
					vs = []astichartjs.DataPoint{}
				}

				// Sometimes the pkt duration time is 0
				if f.PktDurationTime.Duration > 0 {
					var d = astichartjs.DataPoint{
						X: f.BestEffortTimestampTime.Seconds(),
						Y: float64(f.PktSize) / f.PktDurationTime.Seconds() / 1024 * 8,
					}
					vs = append(vs, d)
				}
			}

			// Loop through datasets
			for _, d := range ds {
				c.Data.Datasets = append(c.Data.Datasets, *d)
			}

			// Add chart
			cs = append(cs, c)
		}
	}
	payload = cs
	return
}
