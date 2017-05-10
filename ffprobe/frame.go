package astiffprobe

import (
	"context"
	"strconv"

	"github.com/pkg/errors"
)

// Frame represents a frame
type Frame struct {
	// frame timestamp estimated using various heuristics
	BestEffortTimestamp     int      `json:"best_effort_timestamp"`
	BestEffortTimestampTime Duration `json:"best_effort_timestamp_time"`
	// picture number in bitstream order
	CodedPictureNumber   int    `json:"coded_picture_number"`
	DisplayPictureNumber int    `json:"display_picture_number"`
	Height               int    `json:"height"`
	InterlacedFrame      Bool   `json:"interlaced_frame"`
	KeyFrame             Bool   `json:"key_frame"`
	MediaType            string `json:"media_type"`
	PictType             string `json:"pict_type"`
	PixFmt               string `json:"pix_fmt"`
	// duration of the corresponding packet, expressed in AVStream->time_base units, 0 if unknown
	PktDuration     int      `json:"pkt_duration"`
	PktDurationTime Duration `json:"pkt_duration_time"`
	// DTS copied from the AVPacket that triggered returning this frame
	PktDts     int      `json:"pkt_dts"`
	PktDtsTime Duration `json:"pkt_dts_time"`
	// reordered pos from the last AVPacket that has been input into the decoder
	PktPos int `json:"pkt_pos,string"`
	// PTS copied from the AVPacket that was decoded to produce this frame
	PktPts     int      `json:"pkt_pts"`
	PktPtsTime Duration `json:"pkt_pts_time"`
	// size of the corresponding packet containing the compressed frame
	PktSize           int   `json:"pkt_size,string"`
	RepeatPict        Bool  `json:"repeat_pict"`
	SampleAspectRatio Ratio `json:"sample_aspect_ratio"`
	StreamIndex       int   `json:"stream_index"`
	ToFieldFirst      Bool  `json:"top_field_first"`
	Width             int   `json:"width"`
}

// Frames returns the frames of a stream
func (f *FFProbe) Frames(ctx context.Context, src string, streamIndex int) (fs []Frame, err error) {
	// Execute
	var o Output
	if o, err = f.exec(ctx, f.binaryPath, "-loglevel", "error", "-show_frames", "-select_streams", strconv.Itoa(streamIndex), "-print_format", "json", src); err != nil {
		err = errors.Wrap(err, "executing failed")
		return
	}
	return o.Frames, nil
}
