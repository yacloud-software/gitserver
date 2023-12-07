package java

import (
	"fmt"
	"golang.conradwood.net/gitserver/coderunners/registry"
	"golang.conradwood.net/gitserver/coderunners/utils"
)

func init() {

	registry.Add(&VersionClassRunner{})
}

type VersionClassRunner struct {
}

func (v *VersionClassRunner) GetName() string {
	return "java version class runner"
}
func (v *VersionClassRunner) Run(o *registry.Opts) error {
	files, err := utils.FindFileByName("YaCloudBuildVersion.java")
	if err != nil {
		return err
	}
	r := utils.Replace{}
	r.Add("private static final long BuildID = ", ";", fmt.Sprintf("%d", o.BuildID))
	r.Add("private static final long RepositoryID = ", ";", fmt.Sprintf("%d", o.RepositoryID))
	r.Add("private static final long BuildTimestamp = ", ";", fmt.Sprintf("%d", o.BuildTimestamp))
	r.Add("private static final String ArtefactName = \"", "\";", o.ArtefactName)
	r.Add("private static final String CommitUserID = \"", "\";", o.CommitUserID)
	for _, f := range files {
		fmt.Printf("Updating %s\n", f)
		r.ReplaceInFile(f)
	}
	return nil
}

