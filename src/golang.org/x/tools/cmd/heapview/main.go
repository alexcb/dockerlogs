// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// heapview is a tool for viewing Go heap dumps.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var port = flag.Int("port", 8080, "service port")

var index = `<!DOCTYPE html>
<script src="js/customelements.js"></script>
<script src="js/typescript.js"></script>
<script src="js/moduleloader.js"></script>
<script>
  System.transpiler = 'typescript';
  System.typescriptOptions = {target: ts.ScriptTarget.ES2015};
  System.locate = (load) => load.name + '.ts';
</script>
<script type="module">
  import {main} from './client/main';
  main();
</script>
`

func toolsDir() string {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Println("error: GOPATH not set. Can't find client files")
		os.Exit(1)
	}
	return filepath.Join(filepath.SplitList(gopath)[0], "/src/golang.org/x/tools")
}

var parseFlags = func() {
	flag.Parse()
}

var addHandlers = func() {
	// Directly serve typescript code in client directory for development.
	http.Handle("/client/", http.StripPrefix("/client",
		http.FileServer(http.Dir(filepath.Join(toolsDir(), "cmd/heapview/client")))))

	// Serve typescript.js and moduleloader.js for development.
	http.HandleFunc("/js/typescript.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(toolsDir(), "third_party/typescript/typescript.js"))
	})
	http.HandleFunc("/js/moduleloader.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(toolsDir(), "third_party/moduleloader/moduleloader.js"))
	})
	http.HandleFunc("/js/customelements.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(toolsDir(), "third_party/webcomponents/customelements.js"))
	})

	// Serve index.html using html string above.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, index)
	})
}

var listenAndServe = func() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
}

func main() {
	parseFlags()
	addHandlers()
	listenAndServe()
}
