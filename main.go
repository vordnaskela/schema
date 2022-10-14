package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"log"
	
	"github.com/vordnaskela/schema/ui"
	"github.com/xeipuuv/gojsonschema"
)

type ValidationResult struct {
	Problems []string
}

func main() {
	srv := &http.Server{
		Addr:        ":8888",
		Handler:     router(),
	}
	
	srv.ListenAndServe()
}

func router() http.Handler {
	mux := http.NewServeMux()
	
	
	mux.HandleFunc("/", indexHandler)
	
	staticFS, _ := fs.Sub(ui.StaticFiles, "dist")
	httpFS := http.FileServer(http.FS(staticFS))
	mux.Handle("/static/", httpFS)
	
	mux.HandleFunc("/api/v1/validate", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		
		schema := r.Form.Get("schema")
		document := r.Form.Get("document")
		result, err := validateJSON(schema, document)
		if err != nil {
			log.Fatal(err)
		}
		
		fmt.Fprintln(w, "Schema: ", schema)
		fmt.Fprintln(w, "Document: ", document)
		if len(result.Problems) > 0 {
			for i, problem := range result.Problems {
				fmt.Fprintf(w, "Problem %d: %v", i+1, problem)
			}
		}
	})
	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintln(w, http.StatusText(http.StatusMethodNotAllowed))
		return
	}
	
	if strings.HasPrefix(r.URL.Path, "/api") {
		http.NotFound(w, r)
		return
	}
	
	if r.URL.Path == "/favicon.ico" {
		rawFile, _ := ui.StaticFiles.ReadFile("dist/favicon.ico")
		w.Write(rawFile)
		return
	}
	
	rawFile, _ := ui.StaticFiles.ReadFile("dist/index.html")
	w.Write(rawFile)
}

func validateJSON(schema, document string) (*ValidationResult, error) {
	
	schemaLoader := gojsonschema.NewStringLoader(schema)
	documentLoader := gojsonschema.NewStringLoader(document)
	
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, err
	}
	
	valRes := &ValidationResult{
		Problems: make([]string, 0, len(result.Errors())),
	}
	
	if !result.Valid() {
		for _, rErr := range result.Errors() {
			valRes.Problems = append(valRes.Problems, rErr.String())
		}
	}
	
	return valRes, nil
}