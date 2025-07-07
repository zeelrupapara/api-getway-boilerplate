// Developer: Saif Hamdan
// Date: 20/7/2023

package config

import (
	"os"
	"path"
	"runtime"
)

func InitTestingPath() {
	_, filename, _, _ := runtime.Caller(0)

	elements := append([]string{path.Dir(filename)}, "..")

	dir := path.Join(elements...)
	err := os.Chdir(dir)
	if err != nil {
		panic(err)
	}
}
