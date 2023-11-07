// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package file provides handlers that serve static files for the ozzo routing package.
package file

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	routing "github.com/go-ozzo/ozzo-routing/v2"
)

type Encoding string

const (
	Brotli = Encoding("br")
	GZip   = Encoding("gzip")
)

// PathMap specifies the mapping between URL paths (keys) and file paths (keys).
// The file paths are relative to Options.RootPath
type PathMap map[string]string

// RootPath stores the current working path
var RootPath string

func init() {
	RootPath, _ = os.Getwd()
}

// Server returns a handler that serves the files as the response content.
// The files being served are determined using the current URL path and the specified path map.
// For example, if the path map is {"/css": "/www/css", "/js": "/www/js"} and the current URL path
// "/css/main.css", the file "<working dir>/www/css/main.css" will be served.
// If a URL path matches multiple prefixes in the path map, the most specific prefix will take precedence.
// For example, if the path map contains both "/css" and "/css/img", and the URL path is "/css/img/logo.gif",
// then the path mapped by "/css/img" will be used.
//
// The usage of URL.Paths containing ".." as path element is forbidden, but ".." can be used in file names.
//
//	import (
//	    "log"
//	    "github.com/studio-b12/ozzo-routing"
//	    "github.com/studio-b12/ozzo-routing/file"
//	)
//
//	r := routing.New()
//	r.Get("/*", file.Server(file.PathMap{
//	     "/css": "/ui/dist/css",
//	     "/js": "/ui/dist/js",
//	}))
func Server(pathMap PathMap, opts ...ServerOptions) routing.Handler {
	options := getServerOptions(opts)

	from, to := parsePathMap(pathMap)

	// security measure: limit the files within options.RootPath
	dir := http.Dir(options.RootPath)

	return func(c *routing.Context) error {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			return routing.NewHTTPError(http.StatusMethodNotAllowed)
		}

		if containsDotDot(c.Request.URL.Path) {
			return routing.NewHTTPError(http.StatusBadRequest, "invalid URL path")
		}

		path, found := matchPath(c.Request.URL.Path, from, to)
		if !found || options.Allow != nil && !options.Allow(c, path) {
			return routing.NewHTTPError(http.StatusNotFound)
		}

		var (
			file  http.File
			fstat os.FileInfo
			err   error
			enc   Encoding
		)

		encodings := negotiateEncodings(c, options.Compression)
		dir := compressionDir{dir, encodings}

		if file, enc, err = dir.Open(path); err != nil {
			if options.CatchAllFile != "" {
				return serveFile(c, dir, options.CatchAllFile)
			}
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		}
		defer file.Close()

		if fstat, err = file.Stat(); err != nil {
			return routing.NewHTTPError(http.StatusNotFound, err.Error())
		}

		if fstat.IsDir() {
			if options.IndexFile == "" {
				return routing.NewHTTPError(http.StatusNotFound)
			}
			return serveFile(c, dir, filepath.Join(path, options.IndexFile))
		}

		if enc != "" {
			c.Response.Header().Set("Content-Encoding", string(enc))
		}
		http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
		return nil
	}
}

func serveFile(c *routing.Context, dir compressionDir, path string) error {
	file, enc, err := dir.Open(path)
	if err != nil {
		return routing.NewHTTPError(http.StatusNotFound, err.Error())
	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return routing.NewHTTPError(http.StatusNotFound, err.Error())
	} else if fstat.IsDir() {
		return routing.NewHTTPError(http.StatusNotFound)
	}

	if enc != "" {
		c.Response.Header().Set("Content-Encoding", string(enc))
	}
	http.ServeContent(c.Response, c.Request, path, fstat.ModTime(), file)
	return nil
}

// Content returns a handler that serves the content of the specified file as the response.
// The file to be served can be specified as an absolute file path or a path relative to RootPath (which
// defaults to the current working path).
// If the specified file does not exist, the handler will pass the control to the next available handler.
//
// The usage of URL.Paths containing ".." as path element is forbidden, but ".." can be used in file names.
func Content(path string, opts ...ServerOptions) routing.Handler {
	options := getServerOptions(opts)

	var dir http.Dir
	if filepath.IsAbs(path) {
		dir = http.Dir(path)
		path = ""
	} else {
		dir = http.Dir(options.RootPath)
	}

	return func(c *routing.Context) error {
		if c.Request.Method != "GET" && c.Request.Method != "HEAD" {
			return routing.NewHTTPError(http.StatusMethodNotAllowed)
		}

		if containsDotDot(c.Request.URL.Path) {
			return routing.NewHTTPError(http.StatusBadRequest, "invalid URL path")
		}

		encodings := negotiateEncodings(c, options.Compression)
		dir := compressionDir{dir, encodings}

		return serveFile(c, dir, path)
	}
}

func parsePathMap(pathMap PathMap) (from, to []string) {
	from = make([]string, len(pathMap))
	to = make([]string, len(pathMap))
	n := 0
	for i := range pathMap {
		from[n] = i
		n++
	}
	sort.Strings(from)
	for i, s := range from {
		to[i] = pathMap[s]
	}
	return
}

func matchPath(path string, from, to []string) (string, bool) {
	for i := len(from) - 1; i >= 0; i-- {
		prefix := from[i]
		if strings.HasPrefix(path, prefix) {
			return to[i] + path[len(prefix):], true
		}
	}
	return "", false
}

func negotiateEncodings(c *routing.Context, available []Encoding) []Encoding {
	if len(available) == 0 {
		return nil
	}

	negotioated := make([]Encoding, 0, len(available))

	acceptEncodings := strings.Split(c.Request.Header.Get("Accept-Encoding"), ",")
	for _, availEnc := range available {
		for _, accEnc := range acceptEncodings {
			if string(availEnc) == strings.TrimSpace(strings.ToLower(accEnc)) {
				negotioated = append(negotioated, availEnc)
			}
		}
	}

	return negotioated
}

// Equivalent to containsDotDot() check in http.ServeFile()
func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, isSlashRune) {
		if ent == ".." {
			return true
		}
	}
	return false
}

func isSlashRune(r rune) bool { return r == '/' || r == '\\' }
