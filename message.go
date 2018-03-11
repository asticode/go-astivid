package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"

	"github.com/asticode/go-astichartjs"
	"github.com/asticode/go-astiffmpeg"
	"github.com/asticode/go-astiffprobe"
	"github.com/asticode/go-astilectron"
	"github.com/asticode/go-astilectron-bootstrap"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astitools/ptr"
	"github.com/pkg/errors"
)

func handleMessages(_ *astilectron.Window, m bootstrap.MessageIn) (p interface{}, err error) {
	// Get payload
	switch m.Name {
	case "visualize.bitrate":
		p, err = handleVisualizeBitrate(m)
	case "visualize.psnr":
		p, err = handleVisualizePSNR(m)
	}
	if err != nil {
		p = err.Error()
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
	SourcePath string   `json:"source_path"`
}

func initVisualize(i bootstrap.MessageIn, labelYAxe string) (b Body, cs []astichartjs.Chart, cp *chartColorPicker, err error) {
	// Decode input
	if err = json.Unmarshal(i.Payload, &b); err != nil {
		err = errors.Wrap(err, "decoding input failed")
		return
	}

	// Create color picker
	cp = newChartColorPicker()

	// Max number of datasets
	if len(b.InputPaths) > len(cp.colors) {
		b.InputPaths = b.InputPaths[:len(cp.colors)]
	}

	// Create charts
	// TODO Increase title font size
	// TODO Resize doesn't work on MacOSX
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
							LabelString: labelYAxe,
						},
					},
				},
			},
		},
		Type: astichartjs.ChartTypeLine,
	}}
	return
}

func handleVisualizeBitrate(i bootstrap.MessageIn) (payload interface{}, err error) {
	// Initialize visualization
	b, cs, cp, err := initVisualize(i, "Bitrate (kb/s)")
	if err != nil {
		err = errors.Wrap(err, "initializing visualize failed")
		return
	}

	// Loop through paths
	var m = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	for _, p := range b.InputPaths {
		wg.Add(1)
		go handleVisualizeBitratePath(p, wg, m, cp, cs[0].Data, &err)
	}
	wg.Wait()
	payload = cs
	return
}

func handleVisualizeBitratePath(p string, wg *sync.WaitGroup, m *sync.Mutex, cp *chartColorPicker, csd *astichartjs.Data, err *error) {
	defer wg.Done()

	// Retrieve streams
	var ss []astiffprobe.Stream
	var errRtn error
	astilog.Debugf("Retrieving streams of %s", p)
	if ss, errRtn = ffprobe.Streams(context.Background(), p); errRtn != nil {
		m.Lock()
		*err = errors.Wrapf(errRtn, "retrieving streams of %s failed", p)
		m.Unlock()
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
		if fs, errRtn = ffprobe.Frames(context.Background(), p, s.Index); errRtn != nil {
			m.Lock()
			*err = errors.Wrapf(errRtn, "retrieving frames of stream %d of %s", s.Index, p)
			m.Unlock()
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
		csd.Datasets = append(csd.Datasets, *d)
		m.Unlock()

		// We only process one stream per path
		break
	}
	return
}

func handleVisualizePSNR(i bootstrap.MessageIn) (payload interface{}, err error) {
	// Initialize visualization
	b, cs, cp, err := initVisualize(i, "PSNR")
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

	// Loop through paths
	fi := []astiffmpeg.Input{{Path: b.SourcePath}}
	var fo []astiffmpeg.Output
	fs := make(map[string]string)
	var ks []string
	for idx, p := range b.InputPaths {
		// Add input
		fi = append(fi, astiffmpeg.Input{Path: p})

		// Create temp file
		var f *os.File
		if f, err = ioutil.TempFile(os.TempDir(), "astivid_"); err != nil {
			err = errors.Wrap(err, "creating temp file failed")
			return
		}
		f.Close()
		fs[p] = f.Name()
		ks = append(ks, p)

		// Make sure the temp file is deleted
		defer os.Remove(f.Name())

		// Add output
		fo = append(fo, astiffmpeg.Output{
			Options: &astiffmpeg.OutputOptions{
				Encoding: &astiffmpeg.EncodingOptions{
					ComplexFilters: []astiffmpeg.ComplexFilterOption{
						{
							Filters:       []string{fmt.Sprintf("scale=%d:%d", s.Width, s.Height)},
							InputStreams:  []astiffmpeg.StreamSpecifier{{Index: astiptr.Int(idx + 1)}},
							OutputStreams: []astiffmpeg.StreamSpecifier{{Name: "scaled"}},
						},
						{
							Filters: []string{fmt.Sprintf("psnr=stats_file=%s", f.Name())},
							InputStreams: []astiffmpeg.StreamSpecifier{
								{Index: astiptr.Int(0)},
								{Name: "scaled"},
							},
						},
					},
				},
				Format: "null",
			},
			Path: "-",
		})
	}

	// Execute ffmpeg command
	if err = ffmpeg.Exec(context.Background(), []string{}, astiffmpeg.GlobalOptions{Log: &astiffmpeg.LogOptions{Level: astiffmpeg.LogLevelError}}, fi, fo); err != nil {
		err = errors.Wrap(err, "executing ffmpeg command failed")
		return
	}

	// Loop through log files
	sort.Strings(ks)
	for _, k := range ks {
		// Read file
		var bs []byte
		if bs, err = ioutil.ReadFile(fs[k]); err != nil {
			err = errors.Wrapf(err, "reading file %s failed", fs[k])
			return
		}

		// Create colors
		clr := cp.next()
		if clr == nil {
			break
		}

		// Create dataset
		var d = &astichartjs.Dataset{
			BackgroundColor: clr.Background,
			BorderColor:     clr.Border,
			Label:           filepath.Base(k),
		}

		// Loop through lines
		var vs []astichartjs.DataPoint
		for idx, l := range bytes.Split(bs, []byte("\n")) {
			// Get line values
			mp := make(map[string]float64)
			for _, li := range bytes.Split(l, []byte(" ")) {
				items := bytes.Split(li, []byte(":"))
				if len(items) == 2 {
					v, errParse := strconv.ParseFloat(string(items[1]), 64)
					if errParse == nil {
						mp[string(items[0])] = math.Min(v, 100)
					}
				}
			}

			// Get average PSNR
			psnr, ok := mp["psnr_avg"]
			if !ok {
				continue
			}

			// Compute average
			if idx%int(s.AvgFramerate*2) == 0 && len(vs) > 0 {
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
		cs[0].Data.Datasets = append(cs[0].Data.Datasets, *d)
	}
	payload = cs
	return
}
