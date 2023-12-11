package utils

import (
	"io/ioutil"
	"os"
)

// returns absolute files
func FindFileByName(name string) ([]string, error) {
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	files, err := findFileInDirByName(path, name)
	return files, err
}
func findFileInDirByName(dir, name string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, f := range files {
		cn := f.Name()
		//	fmt.Printf("File: \"%s\"\n", cn)
		if cn == name {
			res = append(res, dir+"/"+cn)
		}
		if f.IsDir() {
			nr, err := findFileInDirByName(dir+"/"+f.Name(), name)
			if err != nil {
				return nil, err
			}
			res = append(res, nr...)
		}
	}
	return res, nil
}


