package main

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
)

//go:embed page.tmpl
var page string

type ByModTime []os.FileInfo

func (a ByModTime) Len() int {
	return len(a)
}

func (a ByModTime) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByModTime) Less(i, j int) bool {
	return a[i].ModTime().After(a[j].ModTime())
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dir, err := os.Open("./")
		if err != nil {
			fmt.Println("Error opening directory:", err)
			return
		}
		defer dir.Close()

		files, err := dir.Readdir(-1)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}

		// Sort the files by modification time in descending order
		sort.Sort(ByModTime(files))

		//parse the embedded template
		tt := template.Must(template.New("page.tmpl").Parse(page))

		// Build the data structure to pass to the template
		data := struct {
			Files []struct {
				Name     string
				Contents string
			}
		}{}
		for _, file := range files {
			if file.IsDir() {
				continue
			}

			// Check if the file has a .txt or .md extension
			if !strings.HasSuffix(file.Name(), ".txt") && !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			var fileName string
			//delete the file extension from the file name
			if strings.HasSuffix(file.Name(), ".txt") {
				fileName = strings.TrimSuffix(file.Name(), ".txt")
			}
			if strings.HasSuffix(file.Name(), ".md") {
				fileName = strings.TrimSuffix(file.Name(), ".md")
			}

			f, err := os.Open(file.Name())
			if err != nil {
				fmt.Println("Error opening file:", err)
				continue
			}

			contents, err := io.ReadAll(f)
			if err != nil {
				fmt.Println("Error reading file:", err)
				continue
			}

			data.Files = append(data.Files, struct {
				Name     string
				Contents string
			}{
				Name:     fileName,
				Contents: string(contents),
			})

			err = f.Close()
			if err != nil {
				fmt.Println("Error closing file:", err)
				continue
			}
		}

		// Render the HTML template and write the output to the response writer
		err = tt.Execute(w, data)
		if err != nil {
			fmt.Println("Error rendering template:", err)
			return
		}
	})

	// Start the HTTP server
	fmt.Println("Listening on http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Error starting HTTP server:", err)
	}
}
