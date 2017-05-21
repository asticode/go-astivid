package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDurationSRT(t *testing.T) {
	d, err := parseDurationSRT("12:34:56")
	assert.EqualError(t, err, "No milliseconds detected in 12:34:56")
	d, err = parseDurationSRT("12:34:56,1234")
	assert.EqualError(t, err, "Invalid number of millisecond digits detected in 12:34:56,1234")
	d, err = parseDurationSRT("12,123")
	assert.EqualError(t, err, "No hours, minutes or seconds detected in 12,123")
	d, err = parseDurationSRT("12:34,123")
	assert.EqualError(t, err, "No hours, minutes or seconds detected in 12:34,123")
	d, err = parseDurationSRT("12:34:56,123")
	assert.NoError(t, err, "")
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+123*time.Millisecond, d)
}

func TestFormatDurationSRT(t *testing.T) {
	s := formatDurationSRT(time.Millisecond)
	assert.Equal(t, "00:00:00,001", s)
	s = formatDurationSRT(10 * time.Millisecond)
	assert.Equal(t, "00:00:00,010", s)
	s = formatDurationSRT(100 * time.Millisecond)
	assert.Equal(t, "00:00:00,100", s)
	s = formatDurationSRT(time.Second + 234*time.Millisecond)
	assert.Equal(t, "00:00:01,234", s)
	s = formatDurationSRT(12*time.Second + 345*time.Millisecond)
	assert.Equal(t, "00:00:12,345", s)
	s = formatDurationSRT(2*time.Minute + 3*time.Second + 456*time.Millisecond)
	assert.Equal(t, "00:02:03,456", s)
	s = formatDurationSRT(20*time.Minute + 34*time.Second + 567*time.Millisecond)
	assert.Equal(t, "00:20:34,567", s)
	s = formatDurationSRT(3*time.Hour + 25*time.Minute + 45*time.Second + 678*time.Millisecond)
	assert.Equal(t, "03:25:45,678", s)
	s = formatDurationSRT(34*time.Hour + 17*time.Minute + 36*time.Second + 789*time.Millisecond)
	assert.Equal(t, "34:17:36,789", s)
}
