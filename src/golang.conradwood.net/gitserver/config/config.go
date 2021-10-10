package config

import (
	"flag"
)

var (
	Gitport = flag.Int("gitport", 5023, "port we listen for the githook on")
)
