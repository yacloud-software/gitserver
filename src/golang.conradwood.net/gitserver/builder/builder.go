package builder

import (
	"fmt"
	//	pb "golang.conradwood.net/apis/gitserver"
	"golang.conradwood.net/gitserver/config"
	"golang.conradwood.net/go-easyops/utils"
	"net"
	"path/filepath"
	"time"
)

const (
	LONG_RUNNING_SECS = 1200 // used to create long running contexts
)

var (
	scriptsdir string
	// files that are required for gitserver to work properly
	requiredFiles = []string{
		"clean-build.sh",
		"go-build.sh",
		"java-build.sh",
		"kicad-build.sh",
		"protos-build.sh",
		"scuserapp-build.sh",
		"dist.sh",
	}
	// location where to look for files
	dirs = []string{
		"/tmp/migrate/",
		"scripts",
	}
	// location where the files where found
	files = make(map[string]string) // short-filename to real filename
)

func Start() error {
	for _, r := range requiredFiles {
		f := ""
		for _, d := range dirs {
			f = findfile(d + "/" + r)
			if f != "" {
				break
			}
		}
		if f == "" {
			return fmt.Errorf("Could not find %s", r)
		}
		files[r] = f
		fmt.Printf("File \"%s\" found at: \"%s\"\n", r, files[r])
		scriptsdir = filepath.Dir(f)
	}
	fmt.Printf("Starting git listener on tcp port %d\n", *config.Gitport)
	go tcpin()
	return nil
}
func tcpin() {
	var err error
	var ln net.Listener
	for {
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", *config.Gitport))
		if err == nil {
			break
		}
		fmt.Printf("Failed to start tcp server (%s)\n", err)
		time.Sleep(time.Duration(1) * time.Second)
	}
	fmt.Printf("Listening on port %d for git\n", *config.Gitport)
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("Accept handle error. weird. (%s)\n", err)
			break
		}
		t := NewTCPConn(conn)
		go t.HandleConnection()
	}
	fmt.Printf("Listener exited\n")
}

func script(name string) string {
	f, found := files[name]
	if !found {
		fmt.Printf("Please list \"%s\" as a required file\n", name)
		panic("unlisted file")
	}
	return f
}

func findfile(name string) string {
	fn, err := utils.FindFile(name)
	if err != nil {
		return ""
	}
	fn, err = filepath.Abs(fn)
	if err != nil {
		fmt.Printf("Failed to fix path: %s\n", err)
		return ""
	}
	return fn
}
