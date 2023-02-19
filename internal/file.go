package internal

import (
	"io"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var (
	imgExtensions = map[string]struct{}{
		".jpeg": {},
		".jpg":  {},
		".png":  {},
	}
)

type fileHandleFunc func(fPath string, dPath string) error

func fileHandlerMove(fPath string, dPath string) error {
	return os.Rename(fPath, dPath)
}

func fileHandlerCopy(fPath string, dPath string) error {
	return fCopy(fPath, dPath)
}

// fCopy copies the contents of the file at srcpath to a regular file at dstpath.
// If dstpath already exists and is not a directory, the function truncates it.
// The function does not copy file modes or file attributes.
// https://stackoverflow.com/a/74107689
func fCopy(srcpath, dstpath string) (err error) {
	r, err := os.Open(srcpath)
	if err != nil {
		return err
	}
	defer r.Close() // ok to ignore error: file was opened read-only.

	w, err := os.Create(dstpath)
	if err != nil {
		return err
	}

	defer func() {
		c := w.Close()
		// Report the error from Close, if any.
		// But do so only if there isn't already
		// an outgoing error.
		if c != nil && err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}

func findImages(path string) ([]string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var ret []string
	for _, f := range files {
		if f.IsDir() {
			continue
		}

		ext := filepath.Ext(f.Name())
		_, isImage := imgExtensions[ext]
		if isImage {
			ret = append(ret, filepath.Join(path, f.Name()))
		}
	}

	return ret, nil
}

func getPath(c *cli.Context) string {
	argPath := c.Args().First()
	if argPath == "" {
		argPath = "."
	}
	return argPath
}
