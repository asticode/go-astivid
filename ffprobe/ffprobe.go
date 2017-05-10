package astiffprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// FFProbe represents an entity capable of running an FFProbe binary
type FFProbe struct {
	binaryPath string
}

// New creates a new FFProbe
func New(c Configuration) *FFProbe {
	return &FFProbe{binaryPath: c.BinaryPath}
}

// exec executes a cmd and returns the unmarshaled output
func (f *FFProbe) exec(ctx context.Context, args ...string) (o Output, err error) {
	// Init
	var cmd = exec.CommandContext(ctx, args[0], args[1:]...)
	var bufOut, bufErr = &bytes.Buffer{}, &bytes.Buffer{}
	cmd.Stdout = bufOut
	cmd.Stderr = bufErr

	// Run cmd
	if err = cmd.Run(); err != nil {
		err = errors.Wrapf(err, "running %s failed with stderr %s", strings.Join(args, " "), bufErr.Bytes())
		return
	}

	// Unmarshal
	if err = json.NewDecoder(bufOut).Decode(&o); err != nil {
		err = errors.Wrapf(err, "unmarshaling %s failed", bufOut.Bytes())
		return
	}
	return
}
