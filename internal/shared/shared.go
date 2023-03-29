package shared

import (
	"flag"

	"github.com/sakirsensoy/genv"
)

type Options struct {
	ListenAddress   string
	ProfilerEnabled bool
}

func InitializeOptions() Options {
	options := Options{}

	flag.StringVar(
		&options.ListenAddress, "listen", genv.Key("LISTEN").Default(":80").String(),
		"Listen address for the web server (:80 by default)",
	)

	profilerEnabled := flag.Bool("profile", false, "Enable profiler web-server on port 6060")

	flag.Parse()

	options.ProfilerEnabled = *profilerEnabled

	return options
}
