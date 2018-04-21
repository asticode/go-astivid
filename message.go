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

	"time"

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

type chartColor struct {
	Background, Border string
}

var chartColors = map[string]chartColor{
	"blue":   {Background: astichartjs.ChartBackgroundColorBlue, Border: astichartjs.ChartBorderColorBlue},
	"green":  {Background: astichartjs.ChartBackgroundColorGreen, Border: astichartjs.ChartBorderColorGreen},
	"red":    {Background: astichartjs.ChartBackgroundColorRed, Border: astichartjs.ChartBorderColorRed},
	"yellow": {Background: astichartjs.ChartBackgroundColorYellow, Border: astichartjs.ChartBorderColorYellow},
	"purple": {Background: astichartjs.ChartBackgroundColorPurple, Border: astichartjs.ChartBorderColorPurple},
	"orange": {Background: astichartjs.ChartBackgroundColorOrange, Border: astichartjs.ChartBorderColorOrange},
}

type Body struct {
	InputPaths map[string]string `json:"input_paths"`
	SourcePath string            `json:"source_path"`
}

func initVisualize(i bootstrap.MessageIn, labelYAxe string) (b Body, cs []astichartjs.Chart, err error) {
	// Decode input
	if err = json.Unmarshal(i.Payload, &b); err != nil {
		err = errors.Wrap(err, "decoding input failed")
		return
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
	b, cs, err := initVisualize(i, "Bitrate (kb/s)")
	if err != nil {
		err = errors.Wrap(err, "initializing visualize failed")
		return
	}

	// Loop through paths
	var m = &sync.Mutex{}
	var wg = &sync.WaitGroup{}
	var ds = make(map[string]astichartjs.Dataset)
	for color, p := range b.InputPaths {
		wg.Add(1)
		go handleVisualizeBitratePath(p, wg, m, color, ds, &err)
	}
	wg.Wait()

	// Order datasets
	var ks []string
	for p := range ds {
		ks = append(ks, p)
		sort.Strings(ks)
	}

	// Add datasets
	for _, k := range ks {
		cs[0].Data.Datasets = append(cs[0].Data.Datasets, ds[k])
	}
	payload = cs
	return
}

func handleVisualizeBitratePath(p string, wg *sync.WaitGroup, m *sync.Mutex, color string, ds map[string]astichartjs.Dataset, err *error) {
	defer wg.Done()

	// Retrieve color
	clr, ok := chartColors[color]
	if !ok {
		return
	}

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

		// Retrieve packets
		var ps []astiffprobe.Packet
		astilog.Debugf("Retrieving packets of stream %d of %s", s.Index, p)
		if ps, errRtn = ffprobe.PacketsOrdered(context.Background(), p, s.Index); errRtn != nil {
			m.Lock()
			*err = errors.Wrapf(errRtn, "retrieving packets of stream %d of %s", s.Index, p)
			m.Unlock()
			return
		}

		// Create dataset
		var d = &astichartjs.Dataset{
			BackgroundColor: clr.Background,
			BorderColor:     clr.Border,
			Label:           filepath.Base(p),
		}

		// Loop through packets
		var vs []astichartjs.DataPoint
		var lastInsertedTime float64
		for _, p := range ps {
			// Since we only analyze video, we assume the framerate is 25 fps by default
			var drt = 40 * time.Millisecond
			if p.DurationTime.Duration > 0 {
				drt = p.DurationTime.Duration
			}

			// Check time
			t := p.PtsTime.Duration
			if len(vs) > 1 && t.Seconds() > lastInsertedTime+2 {
				// Compute sum
				var sum float64
				for _, dp := range vs {
					sum += dp.Y
				}

				// Append
				d.Data = append(d.Data, astichartjs.DataPoint{
					X: vs[0].X,
					Y: sum / float64(len(vs)),
				})

				// Reset
				lastInsertedTime = vs[len(vs)-1].X
				vs = []astichartjs.DataPoint{}
			}

			// Append data point
			var dp = astichartjs.DataPoint{
				X: t.Seconds(),
				Y: float64(p.Size) / drt.Seconds() / 1024 * 8,
			}
			vs = append(vs, dp)
		}

		// Append dataset
		m.Lock()
		ds[p] = *d
		m.Unlock()

		// We only process one stream per path
		break
	}
	return
}

func handleVisualizePSNR(i bootstrap.MessageIn) (payload interface{}, err error) {
	// Initialize visualization
	b, cs, err := initVisualize(i, "PSNR")
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
	type output struct {
		color chartColor
		path  string
	}
	fs := make(map[string]output)
	var ks []string
	for color, p := range b.InputPaths {
		// Retrieve color
		clr, ok := chartColors[color]
		if !ok {
			continue
		}

		// Add input
		fi = append(fi, astiffmpeg.Input{Path: p})

		// Create temp file
		var f *os.File
		if f, err = ioutil.TempFile(os.TempDir(), "astivid_"); err != nil {
			err = errors.Wrap(err, "creating temp file failed")
			return
		}
		f.Close()
		fs[p] = output{
			color: clr,
			path:  f.Name(),
		}
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
							InputStreams:  []astiffmpeg.StreamSpecifier{{Index: astiptr.Int(len(fi) - 1)}},
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
	if err = ffmpeg.Exec(context.Background(), astiffmpeg.GlobalOptions{Log: &astiffmpeg.LogOptions{Level: astiffmpeg.LogLevelError}}, fi, fo); err != nil {
		err = errors.Wrap(err, "executing ffmpeg command failed")
		return
	}

	// Loop through log files
	sort.Strings(ks)
	for _, k := range ks {
		// Read file
		var bs []byte
		if bs, err = ioutil.ReadFile(fs[k].path); err != nil {
			err = errors.Wrapf(err, "reading file %s failed", fs[k])
			return
		}

		// Create dataset
		var d = &astichartjs.Dataset{
			BackgroundColor: fs[k].color.Background,
			BorderColor:     fs[k].color.Border,
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
