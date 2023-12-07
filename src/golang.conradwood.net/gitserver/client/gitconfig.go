package main

import (
	"golang.conradwood.net/go-easyops/utils"
	"strings"
)

type GitConfig struct {
	Sections []*GitSection
}
type GitSection struct {
	Name    string
	Entries map[string]string
}

func ParseGitConfig(fname string) (*GitConfig, error) {
	rf, err := utils.ReadFile(fname)
	if err != nil {
		return nil, err
	}
	var section *GitSection
	res := &GitConfig{}
	for _, line := range strings.Split(string(rf), "\n") {
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sn := strings.TrimPrefix(line, "[")
			sn = strings.TrimSuffix(sn, "]")
			section = &GitSection{Name: sn, Entries: make(map[string]string)}
			res.Sections = append(res.Sections, section)
			continue
		}
		if section == nil {
			continue
		}
		if len(line) < 2 {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		k := strings.Trim(kv[0], " ")
		k = strings.Trim(k, "\t")
		v := strings.Trim(kv[1], " ")
		section.Entries[k] = v

	}

	return res, nil
}

func (g *GitConfig) GetEntry(section, key string) string {
	for _, s := range g.Sections {
		if s.Name != section {
			continue
		}
		return s.Entries[key]
	}
	return ""
}

