package git2

import (
	"strings"
)

type GitURL struct {
	url string
}

func NewGitURL(url string) *GitURL {
	return &GitURL{url: url}
}

/* Examples:
/foobar.git/info/refs?foo=bar -> foobar
/cnw/foobar.git/info/refs?foo=bar -> cnw/foobar
/cnw/foobar/info/refs?foo=bar -> cnw/foobar
*/
func (g *GitURL) RepoPath() string {
	s := strings.TrimPrefix(g.url, "/git/")
	idx := strings.Index(s, ".git/")
	if idx != -1 {
		return strings.Trim(s[:idx+4], "/")
	}
	return "NOTYET"
}
