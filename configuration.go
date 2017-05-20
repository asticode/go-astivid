package main

import (
	"flag"
	"log"

	"github.com/BurntSushi/toml"
	astilog "github.com/asticode/go-astilog"
	"github.com/asticode/go-astivid/ffprobe"
	"github.com/imdario/mergo"
)

// Flags
var (
	configPath    = flag.String("config", "", "the config path")
	pathStatic    = flag.String("path-static", "", "the static path")
	pathTemplates = flag.String("path-templates", "", "the templates path")
	serverAddr    = flag.String("server-addr", "", "the server addr")
)

// Configuration represents a configuration
type Configuration struct {
	FFProbe       astiffprobe.Configuration `toml:"ffprobe"`
	Logger        astilog.Configuration     `toml:"logger"`
	PathStatic    string                    `toml:"path_static"`
	PathTemplates string                    `toml:"path_templates"`
	ServerAddr    string                    `toml:"server_addr"` // Should be of the form host:port
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
		PathStatic:    "resources/static",
		PathTemplates: "resources/templates",
		ServerAddr:    "127.0.0.1:",
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
		FFProbe:       astiffprobe.FlagConfig(),
		Logger:        astilog.FlagConfig(),
		PathStatic:    *pathStatic,
		PathTemplates: *pathTemplates,
		ServerAddr:    *serverAddr,
	}

	// Merge configs
	if err := mergo.Merge(&c, gc); err != nil {
		log.Fatalf("%v while merging configs", err)
	}

	// Return
	return c
}
