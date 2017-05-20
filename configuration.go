package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
	"github.com/imdario/mergo"
)

// Flags
var (
	configPath = flag.String("config", "", "the config path")
)

// Configuration represents a configuration
type Configuration struct {
	FFProbe astiffprobe.Configuration `toml:"ffprobe"`
	Logger  astilog.Configuration     `toml:"logger"`
}

// newConfiguration creates a new configuration object
func newConfiguration() Configuration {
	// Global config
	var gc = Configuration{
		FFProbe: astiffprobe.Configuration{
			BinaryPath: "ffprobe",
		},
		Logger: astilog.Configuration{
			AppName: "astivid",
		},
	}

	// Local config
	if *configPath != "" {
		// Decode local config
		if _, err := toml.DecodeFile(*configPath, &gc); err != nil {
			log.Fatalf("%v while decoding the config path %s", err, *configPath)
		}
	}

	// Flag config
	var c = Configuration{
		FFProbe: astiffprobe.FlagConfig(),
		Logger:  astilog.FlagConfig(),
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		log.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
