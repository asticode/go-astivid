package astiffprobe

import "flag"

// Flags
var (
	BinaryPath = flag.String("ffprobe-binary-path", "", "the FFProbe binary path")
)

// Configuration represents the ffmpeg configuration
type Configuration struct {
	BinaryPath string `toml:"binary_path"`
}

// FlagConfig generates a Configuration based on flags
func FlagConfig() Configuration {
	return Configuration{
		BinaryPath: *BinaryPath,
	}
}
