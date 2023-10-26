package file

import (
	"io/fs"
	"path/filepath"

	routing "github.com/go-ozzo/ozzo-routing/v2"
)

// ServerOptions defines the possible options for the Server handler.
type ServerOptions struct {
	// The path that all files to be served should be located within. The path map passed to the Server method
	// are all relative to this path. This property can be specified as an absolute file path or a path relative
	// to the current working path. If not set, this property defaults to the current working path.
	RootPath string
	// The file (e.g. index.html) to be served when the current request corresponds to a directory.
	// If not set, the handler will return a 404 HTTP error when the request corresponds to a directory.
	// This should only be a file name without the directory part.
	IndexFile string
	// The file to be served when no file or directory matches the current request.
	// If not set, the handler will return a 404 HTTP error when no file/directory matches the request.
	// The path of this file is relative to RootPath
	CatchAllFile string
	// A function that checks if the requested file path is allowed. If allowed, the function
	// may do additional work such as setting Expires HTTP header.
	// The function should return a boolean indicating whether the file should be served or not.
	// If false, a 404 HTTP error will be returned by the handler.
	Allow func(*routing.Context, string) bool
	// Define available compression encodings for serving files. Encodings are negotiated against the
	// unser agent. The first encoding which matches the accepted encodings from the user agent as well
	// as is available as file is served to the client.
	Compression []Encoding
	// The FS to be used to serve files from. When set, this overrides RootPath.
	FS fs.FS
}

// Merge takes another instance of ServerOptions and merges it with the current instance.
// Thereby other overwrites values in t if existent. The merged instance is returned as new
// ServerOptions instance.
func (t ServerOptions) Merge(other ServerOptions) (new ServerOptions) {
	new = t

	if other.Allow != nil {
		new.Allow = other.Allow
	}
	if other.CatchAllFile != "" {
		new.CatchAllFile = other.CatchAllFile
	}
	if other.Compression != nil {
		new.Compression = other.Compression
	}
	if other.IndexFile != "" {
		new.IndexFile = other.IndexFile
	}
	if other.RootPath != "" {
		new.RootPath = other.RootPath
	}
	if other.FS != nil {
		new.FS = other.FS
	}

	return new
}

func getServerOptions(opts []ServerOptions) ServerOptions {
	var options ServerOptions

	for _, opt := range opts {
		options = options.Merge(opt)
	}

	if !filepath.IsAbs(options.RootPath) {
		options.RootPath = filepath.Join(RootPath, options.RootPath)
	}

	return options
}
