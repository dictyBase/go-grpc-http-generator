package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func removeAllTemp(folders ...string) error {
	for _, f := range folders {
		if err := os.RemoveAll(f); err != nil {
			return fmt.Errorf("error in removing folder %s %s", f, err)
		}
	}
	return nil
}

func createValidatorTemp(prefix, path string) (string, string, error) {
	dir, err := ioutil.TempDir(os.TempDir(), prefix)
	if err != nil {
		return "", "", fmt.Errorf("error in creating temp dir %s", err)
	}
	out := filepath.Join(dir, path)
	err = os.MkdirAll(out, 0774)
	if err != nil {
		return "", "", fmt.Errorf("error in creating validator-proto-path %s %s", out, err)
	}
	return dir, out, nil
}
