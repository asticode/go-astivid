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
	"github.com/pkg/errors"
)

// MessageError represents an error message
type MessageError struct {
	Error string `json:"error"`
}

// handleMessages handles messages
func handleMessages(w *astilectron.Window, m bootstrap.MessageIn) (payload interface{}, _ error) {
	// Get payload
	var errPayload error
	switch m.Name {
	case "get.frames":
		payload, errPayload = handleGetFrames(m)
	}

	// Process error
	if errPayload != nil {
		astilog.Error(errPayload)
		payload = MessageError{Error: errPayload.Error()}
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
				Options: astichartjs.Options{
					Scales: astichartjs.Scales{
						XAxes: []astichartjs.Axis{
							{
								Position: astichartjs.ChartAxisPositionsBottom,
								ScaleLabel: astichartjs.ScaleLabel{
									Display:     true,
									LabelString: "Timestamp (s)",
								},
								Type: astichartjs.ChartAxisTypesLinear,
							},
						},
						YAxes: []astichartjs.Axis{
							{
								ScaleLabel: astichartjs.ScaleLabel{
									Display:     true,
									LabelString: "Bitrate (kb/s)",
								},
							},
						},
					},
					Title: astichartjs.Title{Display: true},
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
				"all": {
					BackgroundColor: astichartjs.ChartBackgroundColorGreen,
					BorderColor:     astichartjs.ChartBorderColorGreen,
					Label:           "All frames",
				},
			}
			for _, f := range fs {
				if _, ok := ds[f.PictType]; !ok {
					switch f.PictType {
					case "B":
						ds[f.PictType] = &astichartjs.Dataset{
							BackgroundColor: astichartjs.ChartBackgroundColorBlue,
							BorderColor:     astichartjs.ChartBorderColorBlue,
							Label:           "B frames",
						}
					case "I":
						ds[f.PictType] = &astichartjs.Dataset{
							BackgroundColor: astichartjs.ChartBackgroundColorRed,
							BorderColor:     astichartjs.ChartBorderColorRed,
							Label:           "I frames",
						}
					case "P":
						ds[f.PictType] = &astichartjs.Dataset{
							BackgroundColor: astichartjs.ChartBackgroundColorYellow,
							BorderColor:     astichartjs.ChartBorderColorYellow,
							Label:           "P frames",
						}
					}
				}
				// Sometimes the pkt duration time is 0
				if f.PktDurationTime.Duration > 0 {
					var d = astichartjs.DataPoint{
						X: f.BestEffortTimestampTime.Seconds(),
						Y: float64(f.PktSize) / f.PktDurationTime.Seconds() / 1024,
					}
					ds[f.PictType].Data = append(ds[f.PictType].Data, d)
					ds["all"].Data = append(ds["all"].Data, d)
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
