package testing

import (
	"os"
	"path"
	"runtime"
)

// simple testing helper that changes the directory to the root of the repository
// this allows all files used in test suites to be defined relative to the repository root
func init() {
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
