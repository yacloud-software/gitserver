package registry

import (
	"fmt"
)

var (
	runners []Runner
)

type Runner interface {
	GetName() string
	Run(*Opts) error
}
type Opts struct {
	BuildID        uint64
	BuildTimestamp uint32
	RepositoryID   uint64
	ArtefactName   string
	CommitUserID   string
}

func All() []Runner {
	return runners
}
func Add(r Runner) {
	fmt.Printf("adding %s\n", r.GetName())

	runners = append(runners, r)
}

