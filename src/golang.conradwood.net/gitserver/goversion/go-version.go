package main

// this creates a compileable go file
// with information about the current build
import (
	"flag"
	"fmt"
	"golang.conradwood.net/go-easyops/utils"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	debug = flag.Bool("debug", false, "debug output")
	repo  = flag.String("repo", "", "Repository this code belongs to")
	file  = flag.String("filename", "", "file to update/create")
)

type BuildDef struct {
	Repository   string
	BuildNumber  int
	BuildDate    time.Time
	RepositoryID int
	Commit       string
}

func main() {
	x := flag.Usage
	flag.Usage = func() {
		x()
		fmt.Fprintf(os.Stderr, "\nRepository: \"dc-tools\"\n\n")
	}

	flag.Parse()
	ns := os.Getenv("BUILD_NUMBER")
	if ns == "" {
		fmt.Println("No BUILD_NUMBER environment variable!")
		os.Exit(10)
	}
	num, err := strconv.Atoi(ns)
	utils.Bail("failed to convert BUILD_NUMBER to integer", err)
	bd := BuildDef{
		Repository:  *repo,
		BuildNumber: num,
		BuildDate:   time.Now(),
		Commit:      os.Getenv("COMMIT_ID"),
	}
	ns = os.Getenv("REPOSITORY_ID")
	if ns != "" {
		num, err = strconv.Atoi(ns)
		utils.Bail("invalid repository id", err)
		bd.RepositoryID = num
	}

	if *file != "" {
		bd.buildfile(*file)
	} else {
		files := findFiles()
		if *debug {
			fmt.Printf("Found %d files to update\n", len(files))
		}
		for _, f := range files {
			bd.buildfile(f)
		}
	}
}
func findFiles() []string {
	tests := []string{
		"src/golang.conradwood.net/vendor/golang.conradwood.net/go-easyops/cmdline/appversion.go",
		"src/golang.singingcat.net/vendor/golang.conradwood.net/go-easyops/cmdline/appversion.go",
		"vendor/golang.conradwood.net/go-easyops/cmdline/appversion.go",
	}
	var res []string
	for _, t := range tests {
		if utils.FileExists(t) {
			res = append(res, t)
		}
	}
	files, err := ioutil.ReadDir("src/")
	utils.Bail("failed to read directory", err)
	for _, f := range files {
		for _, t := range tests {
			fname := "src/" + f.Name() + "/" + t
			if *debug {
				fmt.Printf("Looking for file \"%s\"\n", fname)
			}
			if utils.FileExists(fname) {
				res = append(res, fname)
			}
		}
	}
	return res
}

func (b *BuildDef) buildfile(filename string) {
	if filename == "" {
		fmt.Println("No filename!")
		os.Exit(10)
	}
	if *debug {
		fmt.Printf("Updating %s\n", filename)
	}
	sb, err := ioutil.ReadFile(filename)
	utils.Bail("Failed to read file", err)
	s := string(sb)
	// regexes must match original file as well as updated file
	// (a git repository might get patched and re-used)
	bnr := regexp.MustCompile("^(.*BUILD_NUMBER.*=.*) \\d+ (.*)$")
	bdesc := regexp.MustCompile(`^(.*BUILD_DESCRIPTION.*=.*").*(".*)$`)
	bts := regexp.MustCompile("^(.*BUILD_TIMESTAMP.*=.*) \\d+ (.*)$")
	brepoid := regexp.MustCompile("^(.*BUILD_REPOSITORY_ID.*=.*) \\d+ (.*)$")
	brepo := regexp.MustCompile(`^(.*BUILD_REPOSITORY .*=.*").*(".*)$`)
	comm := regexp.MustCompile(`^(.*BUILD_COMMIT.*=.*").*(".*)$`)
	ns := ""
	for _, line := range strings.Split(s, "\n") {

		repl := fmt.Sprintf("${1} %d ${2}", b.BuildNumber)
		line = bnr.ReplaceAllString(line, repl)

		repl = fmt.Sprintf("${1}Build #%d of %s at %s on host %s${2}",
			b.BuildNumber, b.Repository, b.BuildDate, os.Getenv("HOSTNAME"))
		line = bdesc.ReplaceAllString(line, repl)

		repl = fmt.Sprintf("${1} %d ${2}", b.BuildDate.Unix())
		line = bts.ReplaceAllString(line, repl)

		repl = fmt.Sprintf("${1} %d ${2}", b.RepositoryID)
		line = brepoid.ReplaceAllString(line, repl)

		repl = fmt.Sprintf("${1}%s${2}", b.Repository)
		line = brepo.ReplaceAllString(line, repl)

		repl = fmt.Sprintf("${1}%s${2}", b.Commit)
		line = comm.ReplaceAllString(line, repl)

		ns = ns + line + "\n"
	}
	err = ioutil.WriteFile(filename, []byte(ns), 0755)
	utils.Bail("Failed to write file", err)
}


