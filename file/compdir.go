package file

import (
	"fmt"
	"net/http"
)

type compressionDir struct {
	dir       http.Dir
	encodings []Encoding
}

func (t compressionDir) Open(path string) (f http.File, enc Encoding, err error) {
	for _, enc = range t.encodings {
		f, err = t.dir.Open(fmt.Sprintf("%s.%s", path, enc))
		if err == nil {
			return f, enc, nil
		}
	}

	f, err = t.dir.Open(path)
	return f, "", err
}
