package main

import (
	"context"
	"encoding/json"
	"path/filepath"

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

type chartColors struct {
	Background, Border string
}

func allChartColors() []*chartColors {
	return []*chartColors{
		{Background: astichartjs.ChartBackgroundColorBlue, Border: astichartjs.ChartBorderColorBlue},
		{Background: astichartjs.ChartBackgroundColorGreen, Border: astichartjs.ChartBorderColorGreen},
		{Background: astichartjs.ChartBackgroundColorRed, Border: astichartjs.ChartBorderColorRed},
		{Background: astichartjs.ChartBackgroundColorYellow, Border: astichartjs.ChartBorderColorYellow},
		{Background: astichartjs.ChartBackgroundColorPurple, Border: astichartjs.ChartBorderColorPurple},
		{Background: astichartjs.ChartBackgroundColorOrange, Border: astichartjs.ChartBorderColorOrange},
	}
}

type chartColorPicker struct {
	colors []*chartColors
}

func newChartColorPicker() *chartColorPicker {
	return &chartColorPicker{colors: allChartColors()}
}

func (p *chartColorPicker) next() (c *chartColors) {
	if len(p.colors) > 0 {
		c = p.colors[0]
		p.colors = p.colors[1:]
		return
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

	// Create charts
	// TODO Increase title font size
	// TODO Resize doesn't work on MacOSX
	// TODO Disable border on btn click
	var cs = []astichartjs.Chart{{
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
		},
		Type: astichartjs.ChartTypeLine,
	}}

	// Create color picker
	cp := newChartColorPicker()

	// Loop through paths
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

			// Retrieve frames
			var fs []astiffprobe.Frame
			astilog.Debugf("Retrieving frames of stream %d of %s", s.Index, p)
			if fs, err = ffprobe.Frames(context.Background(), p, s.Index); err != nil {
				err = errors.Wrapf(err, "retrieving frames of stream %d of %s", s.Index, p)
				return
			}

			// Create colors
			clr := cp.next()
			if clr == nil {
				continue
			}

			// Create dataset
			var d = &astichartjs.Dataset{
				BackgroundColor: clr.Background,
				BorderColor:     clr.Border,
				Label:           filepath.Base(p),
			}

			// Loop through frames
			var vs []astichartjs.DataPoint
			for _, f := range fs {
				// Compute average
				if f.PictType == "I" && len(vs) > 0 {
					var sum float64
					for _, dp := range vs {
						sum += dp.Y
					}
					d.Data = append(d.Data, astichartjs.DataPoint{
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

			// Append dataset
			cs[0].Data.Datasets = append(cs[0].Data.Datasets, *d)

			// We only process one stream per path
			break
		}
	}
	payload = cs
	return
}
