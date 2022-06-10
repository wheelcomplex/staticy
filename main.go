// a static file server base on net/http
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// https://github.com/jordan-wright/unindexed

// FileSystem is an implementation of a standard http.FileSystem
// without the ability to list files in the directory.
// This implementation is largely inspired by
// https://www.alexedwards.net/blog/disable-http-fileserver-directory-listings
type FileSystem struct {
	fs http.FileSystem
}

// Open returns a file from the static directory. If the requested path ends
// with a slash, there is a check for an index.html file. If none exists, then
// an os.ErrPermission error is returned, causing a 403 Forbidden error to be
// returned to the client
func (ufs FileSystem) Open(name string) (http.File, error) {
	f, err := ufs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	// Check to see if what we opened was a directory. If it was, we will
	// return an error
	s, err := f.Stat()
	if s.IsDir() {
		index := strings.TrimSuffix(name, "/") + "/index.html"
		_, err := ufs.fs.Open(index)
		if err != nil {
			return nil, os.ErrPermission
		}
	}
	return f, nil
}

// unindexedDir is a drop-in replacement for http.Dir, providing an unindexed
// filesystem for serving static files.
func unindexedDir(filepath string) http.FileSystem {
	return FileSystem{
		fs: http.Dir(filepath),
	}
}

//
func serverWithLogHandle(srv http.Handler) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// remote ip host proto Method url
		log.Printf("%s %s %s %s %s\n", r.RemoteAddr, r.Host, r.Proto, r.Method, r.URL)
		srv.ServeHTTP(w, r)
	}

}

func main() {
	log.SetPrefix("staticy ")
	docroot := flag.String("docroot", "./static", "document root")
	hostport := flag.String("listen", "0.0.0.0:8000", "listen to host:port")
	indexing := flag.Bool("indexing", false, "disable directory indexing (default false)")

	flag.Parse()

	*docroot, _ = filepath.Abs(*docroot)

	fileServer := http.FileServer(http.Dir(*docroot))
	if *indexing {
		fileServer = http.FileServer(unindexedDir(*docroot))
	}

	http.HandleFunc("/", serverWithLogHandle(fileServer))

	log.Printf("http static server listen on http://%s for %s\n", *hostport, *docroot)
	log.Printf("directory indexing: %v\n", !*indexing)

	err := http.ListenAndServe(*hostport, nil)
	if err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}

}
