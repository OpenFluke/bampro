package main

import (
	"io/ioutil"
	"strings"
)

func findBundle(dir, prefix, suffix string) string {
	files, _ := ioutil.ReadDir(dir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), prefix) && strings.HasSuffix(f.Name(), suffix) {
			return f.Name()
		}
	}
	return ""
}
