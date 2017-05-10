package astiffprobe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockedFrame = `{
    "media_type": "video",
    "stream_index": 0,
    "key_frame": 1,
    "pkt_pts": 126009,
    "pkt_pts_time": "1.400100",
    "pkt_dts": 126009,
    "pkt_dts_time": "1.400100",
    "best_effort_timestamp": 126009,
    "best_effort_timestamp_time": "1.400100",
    "pkt_duration": 3600,
    "pkt_duration_time": "0.040000",
    "pkt_pos": "940",
    "pkt_size": "72056",
    "width": 1920,
    "height": 1080,
    "pix_fmt": "yuv420p",
    "sample_aspect_ratio": "1:1",
    "pict_type": "I",
    "coded_picture_number": 4,
    "display_picture_number": 0,
    "interlaced_frame": 1,
    "top_field_first": 1,
    "repeat_pict": 0
}`

func TestFFProbe_Frames(t *testing.T) {
	var f Frame
	err := json.Unmarshal([]byte(mockedFrame), &f)
	assert.NoError(t, err)
	assert.Equal(t, Frame{BestEffortTimestamp: 126009, BestEffortTimestampTime: 1400100000, CodedPictureNumber: 4, Height: 1080, InterlacedFrame: true, KeyFrame: true, MediaType: "video", PictType: "I", PixFmt: "yuv420p", PktDuration: 3600, PktDurationTime: 40000000, PktDts: 126009, PktDtsTime: 1400100000, PktPos: 940, PktPts: 126009, PktPtsTime: 1400100000, PktSize: 72056, SampleAspectRatio: Ratio{Width: 1, Height: 1}, ToFieldFirst: true, Width: 1920}, f)
}
