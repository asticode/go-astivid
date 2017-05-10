package astiffprobe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var mockedStreams = `[{
    "index": 0,
    "codec_name": "h264",
    "codec_long_name": "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10",
    "profile": "High",
    "codec_type": "video",
    "codec_time_base": "29/2050",
    "codec_tag_string": "avc1",
    "codec_tag": "0x31637661",
    "width": 1920,
    "height": 1080,
    "coded_width": 1920,
    "coded_height": 1080,
    "has_b_frames": 3,
    "sample_aspect_ratio": "1:1",
    "display_aspect_ratio": "16:9",
    "pix_fmt": "yuv420p",
    "level": 40,
    "color_range": "tv",
    "color_space": "bt709",
    "color_transfer": "bt709",
    "color_primaries": "bt709",
    "chroma_location": "left",
    "refs": 1,
    "is_avc": "true",
    "nal_length_size": "4",
    "r_frame_rate": "50/1",
    "avg_frame_rate": "1025/29",
    "time_base": "1/90000",
    "start_pts": 100980,
    "start_time": "1.122000",
    "duration_ts": 936180,
    "duration": "10.402000",
    "bit_rate": "5015020",
    "bits_per_raw_sample": "8",
    "nb_frames": "328",
    "disposition": {
	"default": 1,
	"dub": 0,
	"original": 0,
	"comment": 0,
	"lyrics": 0,
	"karaoke": 0,
	"forced": 0,
	"hearing_impaired": 0,
	"visual_impaired": 0,
	"clean_effects": 0,
	"attached_pic": 0,
	"timed_thumbnails": 0
    },
    "tags": {
	"language": "und",
	"handler_name": "VideoHandler"
    }
},
{
    "index": 1,
    "codec_name": "mp2",
    "codec_long_name": "MP2 (MPEG audio layer 2)",
    "codec_type": "audio",
    "codec_time_base": "1/48000",
    "codec_tag_string": "mp4a",
    "codec_tag": "0x6134706d",
    "sample_fmt": "s16p",
    "sample_rate": "48000",
    "channels": 2,
    "channel_layout": "stereo",
    "bits_per_sample": 0,
    "r_frame_rate": "0/0",
    "avg_frame_rate": "0/0",
    "time_base": "1/48000",
    "start_pts": 0,
    "start_time": "0.000000",
    "duration_ts": 480384,
    "duration": "10.008000",
    "bit_rate": "192027",
    "max_bit_rate": "192027",
    "nb_frames": "417",
    "disposition": {
	"default": 1,
	"dub": 0,
	"original": 0,
	"comment": 0,
	"lyrics": 0,
	"karaoke": 0,
	"forced": 0,
	"hearing_impaired": 0,
	"visual_impaired": 0,
	"clean_effects": 0,
	"attached_pic": 0,
	"timed_thumbnails": 0
    },
    "tags": {
	"language": "fre",
	"handler_name": "SoundHandler"
    }
}]`

func TestFFProbe_Streams(t *testing.T) {
	var s []Stream
	err := json.Unmarshal([]byte(mockedStreams), &s)
	assert.NoError(t, err)
	assert.Equal(t, []Stream{Stream{AvgFramerate: Division(float64(1025) / float64(29)), Bitrate: 5015020, BitsPerRawSample: 8, ChromaLocation: "left", CodecLongName: "H.264 / AVC / MPEG-4 AVC / MPEG-4 part 10", CodecName: "h264", CodecTag: "0x31637661", CodecTagString: "avc1", CodecTimeBase: "29/2050", CodecType: "video", CodedHeight: 1080, CodedWidth: 1920, ColorPrimaries: "bt709", ColorRange: "tv", ColorSpace: "bt709", ColorTransfer: "bt709", DisplayAspectRatio: Ratio{Width: 16, Height: 9}, Disposition: Disposition{Default: true}, Duration: 10402000000, DurationTs: 936180, HasBFrames: 3, Height: 1080, Index: 0, IsAVC: true, Level: 40, NalLengthSize: 4, NbFrames: 328, PixFmt: "yuv420p", Profile: "High", Refs: 1, RFrameRate: 50, SampleAspectRatio: Ratio{Width: 1, Height: 1}, StartPts: 100980, StartTime: 1122000000, Tags: Tags{HandlerName: "VideoHandler", Language: "und"}, TimeBase: Division(float64(1) / float64(90000)), Width: 1920}, Stream{Bitrate: 192027, ChannelLayout: "stereo", Channels: 2, CodecLongName: "MP2 (MPEG audio layer 2)", CodecName: "mp2", CodecTag: "0x6134706d", CodecTagString: "mp4a", CodecTimeBase: "1/48000", CodecType: "audio", Disposition: Disposition{Default: true}, Duration: 10008000000, DurationTs: 480384, Index: 1, NbFrames: 417, SampleFmt: "s16p", SampleRate: 48000, Tags: Tags{HandlerName: "SoundHandler", Language: "fre"}, TimeBase: Division(float64(1) / float64(48000))}}, s)
}
