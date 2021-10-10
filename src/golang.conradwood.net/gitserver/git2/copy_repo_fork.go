package git2

import (
	"fmt"
	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/db"
	"golang.conradwood.net/go-easyops/authremote"
	"golang.conradwood.net/go-easyops/linux"
	"golang.conradwood.net/go-easyops/utils"
	"strings"
)

var (
	copy_triggers = make(chan *CopyTrigger)
)

type CopyTrigger struct {
	source *pb.SourceRepository
	dest   *pb.SourceRepository
}

func init() {
	go start_copy_thread()
}
func copyrepo(source, dest *pb.SourceRepository) {
	ct := &CopyTrigger{
		source: source,
		dest:   dest,
	}
	copy_triggers <- ct
}
func start_copy_thread() {
	for {
		ct := <-copy_triggers
		ct.Debugf("starting copy...")
		err := ct.Copy()
		if err != nil {
			ct.Errorf("Failed to copy (%s)", utils.ErrorString(err))
		}
	}
}
func (ct *CopyTrigger) Errorf(format string, args ...interface{}) {
	txt := fmt.Sprintf("[copy %d->%d] ", ct.source.ID, ct.dest.ID)
	fmt.Printf(txt+format+"\n", args...)
}
func (ct *CopyTrigger) Debugf(format string, args ...interface{}) {
	txt := fmt.Sprintf("[copy %d->%d] ", ct.source.ID, ct.dest.ID)
	fmt.Printf(txt+format+"\n", args...)
}
func (ct *CopyTrigger) Copy() error {
	// refresh copy to check if it has stopped 'forking' meanwhile
	ctx := authremote.Context()
	fr, err := db.NewDBSourceRepository(psql).ByID(ctx, ct.dest.ID)
	if err != nil {
		return err
	}
	ct.dest = fr
	if !ct.dest.Forking {
		ct.Debugf("no longer forking")
		return nil
	}
	linux.New()
	src := *root_dir + "/" + strings.Trim(ct.source.FilePath, "/")
	dest := *root_dir + "/" + strings.Trim(ct.dest.FilePath, "/")
	ct.Debugf("Copying \"%s\" to \"%s\"...", src, dest)
	err = linux.CopyDir(src, dest)
	if err != nil {
		return err
	}
	// now update database to say it is no longer forking
	ctx = authremote.Context()
	fr, err = db.NewDBSourceRepository(psql).ByID(ctx, ct.dest.ID)
	if err != nil {
		return err
	}
	fr.Forking = false
	err = db.NewDBSourceRepository(psql).Update(ctx, fr)
	if err != nil {
		return err
	}
	ct.Debugf("Copy complete and database updated")
	return nil
}
