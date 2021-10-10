package utils

import (
	"fmt"
	cu "golang.conradwood.net/go-easyops/utils"
	"strings"
)

type Replace struct {
	patterns []*replacePattern
}
type replacePattern struct {
	first   string
	end     string
	replace string
}

func (r *Replace) ReplaceInFile(filename string) error {
	ct, err := cu.ReadFile(filename)
	if err != nil {
		return err
	}
	sct := string(ct)
	res := ""
	for _, line := range strings.Split(sct, "\n") {
		line = r.ReplaceLine(line)
		res = res + line + "\n"
	}
	err = cu.WriteFile(filename, []byte(res))
	if err != nil {
		return err
	}
	return nil
}

func (r *Replace) ReplaceLine(line string) string {
	for _, p := range r.patterns {
		fidx := strings.Index(line, p.first)
		lidx := strings.Index(line, p.end)
		if fidx == -1 || lidx == -1 {
			continue
		}
		fidx = fidx + len(p.first)
		l1 := line[:fidx]
		l2 := line[lidx:]
		nl := l1 + p.replace + l2
		fmt.Printf("Replacing \"%s\" with \"%s\"\n", line, nl)
		return nl

	}
	return line
}

func (r *Replace) Add(first string, end string, replace string) {
	rp := &replacePattern{first: first, end: end, replace: replace}
	r.patterns = append(r.patterns, rp)
}
