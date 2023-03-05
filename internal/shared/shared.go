package shared

import (
	"flag"

	"github.com/sakirsensoy/genv"
)

type Options struct {
	ListenAddress string
}

func InitializeOptions() Options {
	options := Options{}

	flag.StringVar(
		&options.ListenAddress, "listen", genv.Key("LISTEN").Default(":80").String(),
		"Listen address for the web server (:80 by default)",
	)

	flag.Parse()

    return options
}
