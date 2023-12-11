package git2

import (
	"context"
	"fmt"
	gitpb "golang.conradwood.net/apis/gitserver"
)

// return true if sourcerepo matches filter
func filterMatch(ctx context.Context, repo *gitpb.SourceRepository, filter *gitpb.RepoFilter) (bool, error) {
	if filter.Tags == nil {
		return true, nil
	}
	mask := uint32((1 << filter.Tags.Tag))
	if (repo.Tags & mask) == 0 {
		if *debug {
			fmt.Printf("Tag: %d, Mask: %08x, Repo: %d\n", filter.Tags.Tag, mask, repo.ID)
		}
		return false, nil
	}
	return true, nil
}


