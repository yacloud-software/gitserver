package config

import (
	"flag"
)

var (
	Gitroot = flag.String("git2_git_dir", "/srv/git", "top level directory underwhich git directories are stored. e.g. /srv/git2")
	Gitport = flag.Int("gitport", 5023, "port we listen for the githook on")
)

