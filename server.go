package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
	"github.com/julienschmidt/httprouter"
	"github.com/pkg/errors"
)

// adaptRouter adapts the router
func adaptRouter(r *httprouter.Router) {
	r.POST("/api/frames", handleAPIFrames)
}

// BodyPaths represents a body containing a list of paths
type BodyPaths struct {
	Paths []string `json:"paths"`
}

// BodyCharts represents a body containing charts
type BodyCharts struct {
	Charts []BodyChart `json:"charts"`
}

// Chart types
const (
	chartTypeLine = "line"
)

// BodyChart represents a body containing a chart
type BodyChart struct {
	Data    BodyChartParentData `json:"data"`
	Options BodyChartOptions    `json:"options"`
	Type    string              `json:"type"`
}

// BodyChartParentData represents a body containing chart parent data
type BodyChartParentData struct {
	Datasets []BodyChartDataset `json:"datasets"`
}

// BodyChartDataset represents a body containing a chart dataset
type BodyChartDataset struct {
	BackgroundColor string          `json:"backgroundColor"`
	BorderColor     string          `json:"borderColor"`
	Data            []BodyChartData `json:"data"`
	Label           string          `json:"label"`
}

// BodyChartData represents a body containing chart data
type BodyChartData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// BodyChartOptions represents a body containing chart options
type BodyChartOptions struct {
	Responsive bool                   `json:"responsive"`
	Scales     BodyChartScalesOptions `json:"scales"`
	Title      BodyChartTitleOptions  `json:"title"`
}

// BodyChartScalesOptions represents a body containing chart scales options
type BodyChartScalesOptions struct {
	XAxes []BodyChartAxisOptions `json:"xAxes"`
	YAxes []BodyChartAxisOptions `json:"yAxes"`
}

// Chart axis positions
const (
	chartAxisPositionsBottom = "bottom"
)

// Chart axis type
const (
	chartAxisTypesLinear = "linear"
)

// BodyChartAxisOptions represents a body containing chart axis options
type BodyChartAxisOptions struct {
	Position   string                     `json:"position,omitempty"`
	ScaleLabel BodyChartScaleLabelOptions `json:"scaleLabel"`
	Type       string                     `json:"type,omitempty"`
}

// BodyChartAxisOptions represents a body containing chart scale label options
type BodyChartScaleLabelOptions struct {
	Display     bool   `json:"display"`
	LabelString string `json:"labelString"`
}

// BodyChartTitleOptions represents a body containing chart title options
type BodyChartTitleOptions struct {
	Display bool   `json:"display"`
	Text    string `json:"text"`
}

// Chart background colors
const (
	chartBackgroundColorBlue   = "rgba(54, 162, 235, 0.2)"
	chartBackgroundColorGreen  = "rgba(75, 192, 192, 0.2)"
	chartBackgroundColororange = "rgba(255, 159, 64, 0.2)"
	chartBackgroundColorPurple = "rgba(153, 102, 255, 0.2)"
	chartBackgroundColorRed    = "rgba(255, 99, 132, 0.2)"
	chartBackgroundColorYellow = "rgba(255, 206, 86, 0.2)"
)

// Chart border colors
const (
	chartBorderColorBlue   = "rgba(54, 162, 235, 1)"
	chartBorderColorGreen  = "rgba(75, 192, 192, 1)"
	chartBorderColorOrange = "rgba(255, 159, 64, 1)"
	chartBorderColorPurple = "rgba(153, 102, 255, 1)"
	chartBorderColorRed    = "rgba(255,99,132,1)"
	chartBorderColorYellow = "rgba(255, 206, 86, 1)"
)

// processErrors processes errors
func processErrors(rw http.ResponseWriter, err *error) {
	if *err != nil {
		astilog.Error(*err)
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(errors.Cause(*err).Error()))
	}
}

// handleAPIFrames handles the /api/frames POST request
func handleAPIFrames(rw http.ResponseWriter, r *http.Request, p httprouter.Params) {
	// Process errors
	var err error
	defer processErrors(rw, &err)

	// Decode input
	var ps BodyPaths
	if err = json.NewDecoder(r.Body).Decode(&ps); err != nil {
		err = errors.Wrap(err, "decoding input failed")
		return
	}

	// Loop through paths
	var cs = BodyCharts{}
	for _, p := range ps.Paths {
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
			var c = BodyChart{
				Options: BodyChartOptions{
					Scales: BodyChartScalesOptions{
						XAxes: []BodyChartAxisOptions{
							{
								Position: chartAxisPositionsBottom,
								ScaleLabel: BodyChartScaleLabelOptions{
									Display:     true,
									LabelString: "Frames index",
								},
								Type: chartAxisTypesLinear,
							},
						},
						YAxes: []BodyChartAxisOptions{
							{
								ScaleLabel: BodyChartScaleLabelOptions{
									Display:     true,
									LabelString: "Bitrate (kb/s)",
								},
							},
						},
					},
					Title: BodyChartTitleOptions{Display: true},
				},
				Type: chartTypeLine,
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
			var ds = make(map[string]*BodyChartDataset)
			for _, f := range fs {
				if _, ok := ds[f.PictType]; !ok {
					switch f.PictType {
					case "B":
						ds[f.PictType] = &BodyChartDataset{
							BackgroundColor: chartBackgroundColorBlue,
							BorderColor:     chartBorderColorBlue,
							Label:           "B frames",
						}
					case "I":
						ds[f.PictType] = &BodyChartDataset{
							BackgroundColor: chartBackgroundColorRed,
							BorderColor:     chartBorderColorRed,
							Label:           "I frames",
						}
					case "P":
						ds[f.PictType] = &BodyChartDataset{
							BackgroundColor: chartBackgroundColorYellow,
							BorderColor:     chartBorderColorYellow,
							Label:           "P frames",
						}
					}
				}
				ds[f.PictType].Data = append(ds[f.PictType].Data, BodyChartData{
					X: f.BestEffortTimestampTime.Seconds(),
					Y: float64(f.PktSize) / f.PktDurationTime.Seconds() / 1024,
				})
			}

			// Loop through datasets
			for _, d := range ds {
				c.Data.Datasets = append(c.Data.Datasets, *d)
			}

			// Add chart
			cs.Charts = append(cs.Charts, c)
		}
	}

	// Write
	if err = json.NewEncoder(rw).Encode(cs); err != nil {
		err = errors.Wrap(err, "writing outputfailed")
		return
	}
}
