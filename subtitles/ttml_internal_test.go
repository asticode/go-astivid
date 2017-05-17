package astisub

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTTMLDuration(t *testing.T) {
	// Unmarshal hh:mm:ss.mmm format
	d, err := newTTLMDuration("12:34:56.789")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+789*time.Millisecond, d.d)

	// Marshal
	b, err := d.MarshalText()
	assert.NoError(t, err)
	assert.Equal(t, "12:34:56.789", string(b))

	// Duration
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+789*time.Millisecond, d.Duration())

	// Unmarshal hh:mm:ss:fff format
	d, err = newTTLMDuration("12:34:56:2")
	assert.NoError(t, err)
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second, d.d)
	assert.Equal(t, 2, d.frames)

	// Duration
	d.framerate = 8
	assert.Equal(t, 12*time.Hour+34*time.Minute+56*time.Second+250*time.Millisecond, d.Duration())
}
