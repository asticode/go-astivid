package main

import (
	"context"
	"encoding/json"
	"path/filepath"

	"sync"

	"io/ioutil"
	"os"

	"bytes"

	"github.com/asticode/go-astichartjs"
	"github.com/asticode/go-astiffprobe"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/ptr"
	"github.com/pkg/errors"
)

func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (_ interface{}, _ error) {
	// Get payload
	switch m.Name {
	case "visualize.bitrate":
		return handleVisualizeBitrate(m)
	case "visualize.psnr":
		return handleVisualizePSNR(m)
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

type Body struct {
	InputPaths []string `json:"input_paths"`
	SourcePath string   `json:"source_paths"`
}

func initVisualize(i bootstrap.MessageIn) (b Body, cs []astichartjs.Chart, cp *chartColorPicker, err error) {
	// Decode input
	if err = json.Unmarshal(i.Payload, &b); err != nil {
		err = errors.Wrap(err, "decoding input failed")
		return
	}

	// Create charts
	// TODO Increase title font size
	// TODO Resize doesn't work on MacOSX
	// TODO Disable border on btn click
	cs = []astichartjs.Chart{{
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
	cp = newChartColorPicker()
	return
}

func handleVisualizeBitrate(i bootstrap.MessageIn) (payload interface{}, err error) {
	b, cs, cp, err := initVisualize(i)
	if err != nil {
		err = errors.Wrap(err, "initializing visualize failed")
		return
	}

	// Loop through paths
	var m = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	for _, p := range b.InputPaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

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
				m.Lock()
				cs[0].Data.Datasets = append(cs[0].Data.Datasets, *d)
				m.Unlock()

				// We only process one stream per path
				break
			}
		}(p)
	}
	wg.Wait()
	payload = cs
	return
}

func handleVisualizePSNR(i bootstrap.MessageIn) (payload interface{}, err error) {
	b, cs, cp, err := initVisualize(i)
	if err != nil {
		err = errors.Wrap(err, "initializing visualize failed")
		return
	}
	// Retrieve source streams
	var ss []astiffprobe.Stream
	astilog.Debugf("Retrieving streams of %s", b.SourcePath)
	if ss, err = ffprobe.Streams(context.Background(), b.SourcePath); err != nil {
		err = errors.Wrapf(err, "retrieving streams of %s failed", b.SourcePath)
		return
	}

	// Get video stream
	var s *astiffprobe.Stream
	for _, v := range ss {
		if v.CodecType == astiffprobe.CodecTypeVideo {
			s = &astiffprobe.Stream{}
			*s = v
			break
		}
	}

	// No video stream
	if s == nil {
		err = errors.New("no video stream in source")
		return
	}

	// Compute number of frames per point
	framesPerPoint := int(s.AvgFramerate * 2)

	// Loop through paths
	var m = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	for _, p := range b.InputPaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// Create temp file
			f, errRtn := ioutil.TempFile(os.TempDir(), "astivid_")
			if errRtn != nil {
				m.Lock()
				err = errors.Wrap(errRtn, "creating temp file failed")
				m.Unlock()
				return
			}
			f.Close()

			// Make sure the temp file is deleted
			defer os.Remove(f.Name())

			// TODO ffmpeg  -i tf1_snow_3min.ts -i stream1.mp4  -lavfi ‘[1]scale=1920:1080[a];[0][a]psnr=stats_file=psnrf_stream1.log’ -f null -

			// Open file
			bs, errRtn := ioutil.ReadFile(f.Name())
			if errRtn != nil {
				m.Lock()
				err = errors.Wrapf(errRtn, "reading file %s failed", f.Name())
				m.Unlock()
				return
			}

			// Create colors
			clr := cp.next()
			if clr == nil {
				return
			}

			// Create dataset
			var d = &astichartjs.Dataset{
				BackgroundColor: clr.Background,
				BorderColor:     clr.Border,
				Label:           filepath.Base(p),
			}

			// Loop through lines
			var vs []astichartjs.DataPoint
			for idx, l := range bytes.Split(bs, []byte("\n")) {
				// TODO Get frame average PSNR
				var psnr float64

				// Compute average
				if idx%framesPerPoint == 0 && len(vs) > 0 {
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

				// Append data point
				vs = append(vs, astichartjs.DataPoint{
					X: float64(idx) / float64(s.AvgFramerate),
					Y: psnr,
				})
			}

			// Append dataset
			m.Lock()
			cs[0].Data.Datasets = append(cs[0].Data.Datasets, *d)
			m.Unlock()
		}(p)
	}
	wg.Wait()
	payload = cs
	return
}
