package filesystem

import (
	"net/http"
	"path/filepath"
)

// All requests for directories (with no index.html file) return a 404 Not Found response, instead of a directory listing or a redirect.
// This works for requests both with and without a trailing slash.
type CustomFileSystem struct {
	Fs http.FileSystem
}

// We Stat() the requested file path and use the IsDir() method to check whether it's a directory or not.
// If it is a directory, we then try to Open() any index.html file in it. If no index.html file exists,
// then this will return a os.ErrNotExist error (which in turn we return and it will be transformed into
// a 404 Not Found response by http.Fileserver). We also call Close() on the original file to avoid a file descriptor leak.
// Otherwise, we just return the file and let http.FileServer do its thing.
func (cfs CustomFileSystem) Open(path string) (http.File, error) {
	file, err := cfs.Fs.Open(path)
	if err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	if fileInfo.IsDir() {
		index := filepath.Join(path, "index.html")

		if _, err := cfs.Fs.Open(index); err != nil {
			closeErr := file.Close()

			if closeErr != nil {
				return nil, closeErr
			}

			return nil, err
		}
	}

	return file, nil
}
