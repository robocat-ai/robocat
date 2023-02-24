package shared

import (
	"flag"

	"github.com/sakirsensoy/genv"
)

type Options struct {
	ListenAddress string

	Profile string
	// "/run.html" by default
	AutostartFile string
	// "/home/robocat/Desktop/uivision" by default
	UIVisionDir string

	Macro string
}

// func getUIVisionDir() string {
// 	uivisionDir := flag.Arg(0)
// 	if len(uivisionDir) == 0 {
// 		uivisionDir = "."
// 	}

// 	uivisionDir, err := filepath.Abs(uivisionDir)
// 	if err != nil {
// 		log.Fatalf("Path '%s' is not valid", uivisionDir)
// 	}

// 	log.Debugf("Got directory for UI.Vision RPA: %s", uivisionDir)

// 	return uivisionDir
// }

func InitializeOptions() Options {
	options := Options{}

	flag.StringVar(
		&options.ListenAddress, "listen", genv.Key("LISTEN").Default(":80").String(),
		"Listen address for the web server (:80 by default)",
	)

	flag.Parse()

	// options.Macro = genv.Key("MACRO").String()
	// if len(options.Macro) == 0 {
	// 	log.Fatal("MACRO must not be empty")
	// }

	return options
}
