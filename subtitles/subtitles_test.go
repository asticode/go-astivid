package astisub_test

import (
	"testing"
	"time"

	"github.com/asticode/go-astivid/subtitles"
	"github.com/stretchr/testify/assert"
)

func assertSubtitles(t *testing.T, i *astisub.Subtitles) {
	// Assert items
	assert.Len(t, i.Items, 6)
	assert.Equal(t, time.Minute+39*time.Second, i.Items[0].StartAt)
	assert.Equal(t, time.Minute+41*time.Second+370*time.Millisecond, i.Items[0].EndAt)
	assert.Equal(t, []string{"(deep rumbling)"}, i.Items[0].Text)
	assert.Equal(t, 2*time.Minute+4*time.Second+200*time.Millisecond, i.Items[1].StartAt)
	assert.Equal(t, 2*time.Minute+7*time.Second+566*time.Millisecond, i.Items[1].EndAt)
	assert.Equal(t, []string{"MAN:", "How did we end up here?"}, i.Items[1].Text)
	assert.Equal(t, 2*time.Minute+12*time.Second+904*time.Millisecond, i.Items[2].StartAt)
	assert.Equal(t, 2*time.Minute+15*time.Second+407*time.Millisecond, i.Items[2].EndAt)
	assert.Equal(t, []string{"This place is horrible."}, i.Items[2].Text)
	assert.Equal(t, 2*time.Minute+20*time.Second+646*time.Millisecond, i.Items[3].StartAt)
	assert.Equal(t, 2*time.Minute+22*time.Second+848*time.Millisecond, i.Items[3].EndAt)
	assert.Equal(t, []string{"Smells like balls."}, i.Items[3].Text)
	assert.Equal(t, 2*time.Minute+28*time.Second+587*time.Millisecond, i.Items[4].StartAt)
	assert.Equal(t, 2*time.Minute+31*time.Second+23*time.Millisecond, i.Items[4].EndAt)
	assert.Equal(t, []string{"We don't belong", "in this shithole."}, i.Items[4].Text)
	assert.Equal(t, 2*time.Minute+31*time.Second+56*time.Millisecond, i.Items[5].StartAt)
	assert.Equal(t, 2*time.Minute+33*time.Second+225*time.Millisecond, i.Items[5].EndAt)
	assert.Equal(t, []string{"(computer playing", "electronic melody)"}, i.Items[5].Text)
}

func mockSubtitles() *astisub.Subtitles {
	return &astisub.Subtitles{Items: []*astisub.Subtitle{{EndAt: 3 * time.Second, StartAt: time.Second, Text: []string{"subtitle-1"}}, {EndAt: 7 * time.Second, StartAt: 3 * time.Second, Text: []string{"subtitle-2"}}}}
}

func TestSubtitles_Add(t *testing.T) {
	var s = mockSubtitles()
	s.Add(time.Second)
	assert.Len(t, s.Items, 2)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
	assert.Equal(t, 2*time.Second, s.Items[0].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[0].EndAt)
}

func TestSubtitles_Duration(t *testing.T) {
	assert.Equal(t, time.Duration(0), astisub.Subtitles{}.Duration())
	assert.Equal(t, 7*time.Second, mockSubtitles().Duration())
}

func TestSubtitles_IsEmpty(t *testing.T) {
	assert.True(t, astisub.Subtitles{}.IsEmpty())
	assert.False(t, mockSubtitles().IsEmpty())
}

func TestSubtitles_Fragment(t *testing.T) {
	var s = mockSubtitles()
	s.Fragment(2 * time.Second)
	assert.Len(t, s.Items, 5)
	assert.Equal(t, time.Second, s.Items[0].StartAt)
	assert.Equal(t, 2*time.Second, s.Items[0].EndAt)
	assert.Equal(t, []string{"subtitle-1"}, s.Items[0].Text)
	assert.Equal(t, 2*time.Second, s.Items[1].StartAt)
	assert.Equal(t, 3*time.Second, s.Items[1].EndAt)
	assert.Equal(t, []string{"subtitle-1"}, s.Items[1].Text)
	assert.Equal(t, 3*time.Second, s.Items[2].StartAt)
	assert.Equal(t, 4*time.Second, s.Items[2].EndAt)
	assert.Equal(t, []string{"subtitle-2"}, s.Items[2].Text)
	assert.Equal(t, 4*time.Second, s.Items[3].StartAt)
	assert.Equal(t, 6*time.Second, s.Items[3].EndAt)
	assert.Equal(t, []string{"subtitle-2"}, s.Items[3].Text)
	assert.Equal(t, 6*time.Second, s.Items[4].StartAt)
	assert.Equal(t, 7*time.Second, s.Items[4].EndAt)
	assert.Equal(t, []string{"subtitle-2"}, s.Items[4].Text)
}

func TestSubtitles_Merge(t *testing.T) {
	var s1 = astisub.Subtitles{Items: []*astisub.Subtitle{{EndAt: 3 * time.Second, StartAt: time.Second}, {EndAt: 8 * time.Second, StartAt: 5 * time.Second}, {EndAt: 12 * time.Second, StartAt: 10 * time.Second}}}
	var s2 = astisub.Subtitles{Items: []*astisub.Subtitle{{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, {EndAt: 7 * time.Second, StartAt: 6 * time.Second}, {EndAt: 11 * time.Second, StartAt: 9 * time.Second}, {EndAt: 14 * time.Second, StartAt: 13 * time.Second}}}
	s1.Merge(s2)
	assert.Len(t, s1.Items, 7)
	assert.Equal(t, &astisub.Subtitle{EndAt: 3 * time.Second, StartAt: time.Second}, s1.Items[0])
	assert.Equal(t, &astisub.Subtitle{EndAt: 4 * time.Second, StartAt: 2 * time.Second}, s1.Items[1])
	assert.Equal(t, &astisub.Subtitle{EndAt: 8 * time.Second, StartAt: 5 * time.Second}, s1.Items[2])
	assert.Equal(t, &astisub.Subtitle{EndAt: 7 * time.Second, StartAt: 6 * time.Second}, s1.Items[3])
	assert.Equal(t, &astisub.Subtitle{EndAt: 11 * time.Second, StartAt: 9 * time.Second}, s1.Items[4])
	assert.Equal(t, &astisub.Subtitle{EndAt: 12 * time.Second, StartAt: 10 * time.Second}, s1.Items[5])
	assert.Equal(t, &astisub.Subtitle{EndAt: 14 * time.Second, StartAt: 13 * time.Second}, s1.Items[6])
}

func TestSubtitles_ForceDuration(t *testing.T) {
	var s = mockSubtitles()
	s.ForceDuration(10 * time.Second)
	assert.Len(t, s.Items, 3)
	assert.Equal(t, 10*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 10*time.Second, s.Items[2].StartAt)
	assert.Equal(t, "...", s.Items[2].Text[0])
	s.Items[2].StartAt = 7 * time.Second
	s.Items[2].EndAt = 12 * time.Second
	s.ForceDuration(10 * time.Second)
	assert.Len(t, s.Items, 3)
	assert.Equal(t, 10*time.Second, s.Items[2].EndAt)
	assert.Equal(t, 7*time.Second, s.Items[2].StartAt)
}
