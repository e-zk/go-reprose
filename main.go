package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	ListenAddr      = "127.0.0.1:8082"
	ReposFile       = "/etc/reprose.txt"
	Title           = "zakaria's golang repo"
	Host            = "go.zakaria.org"
	DefaultRedirect = "https://git.zakaria.org/"
)

var Repos map[string]string

func readRepos() error {
	Repos = make(map[string]string)
	f, err := os.Open(ReposFile)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		split := strings.Fields(line)
		if len(split) != 2 {
			return fmt.Errorf("error parsing repos file")
		}
		Repos[split[0]] = split[1]
	}
	return nil
}

func isGoGet(req *http.Request) bool {
	q := req.URL.Query()
	goget := q.Get("go-get")
	if len(goget) == 0 {
		return false
	}
	if goget == "1" {
		return true
	}
	return false
}

func printRepo(repo, path string) string {
	var b strings.Builder
	// fmt.Fprintf(&b, "<a href=\"%s/%s\"><img src=\"%s/%s?status.svg\" alt=\"godocs.io\"/></a>", GoDocsRoot, path, GoDocsRoot, path)
	//fmt.Fprintf(&b, "[<a href=\"%s/%s\">docs</a>] ", GoDocsRoot, path)
	fmt.Fprintf(&b, "<a href=\"/%s\" title=\"%s/%s -> %s\">%s/%s</a> ", repo, Host, repo, path, Host, repo)
	return b.String()
}

func main() {
	readRepos()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<pre>%s\n\n", Title)
			fmt.Fprintf(w, "Repos:\n")
			for k, v := range Repos {
				//fmt.Fprintf(w, " — <a href=\"/%s\">%s/%s</a>\n", k, Host, k)
				fmt.Fprintf(w, " — %s\n", printRepo(k, v))
			}
			fmt.Fprintf(w, "</pre>")

			return
		}

		for k, v := range Repos {
			if req.URL.Path == "/"+k {
				if !isGoGet(req) {
					http.Redirect(w, req, v, http.StatusPermanentRedirect)
					return
				}
				fmt.Fprintf(w, "<html><head>")
				fmt.Fprintf(w, "<title>%s</title>", k)
				fmt.Fprintf(w, "<meta name=\"go-import\" content=\"%s/%s git %s\">", Host, k, v)
				fmt.Fprintf(w, "<meta name=\"go-source\" content=\"%s/%s _ %s/tree{/dir} %s/tree{/dir}/{file}#n{line}\">", Host, k, v, v)
				fmt.Fprintf(w, "</head>")
				fmt.Fprintf(w, "<body>go get %s/%s</body>", Host, k)
				fmt.Fprintf(w, "</html>")
				return
			}
		}
	})
	http.ListenAndServe("127.0.0.1:8082", mux)
}
