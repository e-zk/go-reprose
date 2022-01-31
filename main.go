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
	Head            = `<meta charset="utf-8"><meta content="width=device-width,initial-scale=1" name="viewport"><style>
	                   pre { line-height: calc(100% * 1.1618); font-size: 100% }
	                   body { margin: 1rem auto; max-width: 50rem; font-size: 100%; line-height: calc(100% * 1.618) }
	                   </style>`
)

type Repo struct {
	Git  string
	Http string
}

var Repos map[string]Repo

func readRepos() error {
	Repos = make(map[string]Repo)
	f, err := os.Open(ReposFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		// skip comments
		if strings.HasPrefix(line, "#") {
			continue
		}

		// split on whitespace
		split := strings.Fields(line)
		if len(split) < 2 {
			return fmt.Errorf("error parsing repos file")
		}

		// if separate http redirect is specified, use it
		// otherwise, use the git url for redirect
		if len(split) == 3 {
			Repos[split[0]] = Repo{split[1], split[2]}
		} else {
			Repos[split[0]] = Repo{split[1], split[1]}
		}
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

func printRepo(base string, repo Repo) string {
	return fmt.Sprintf("<a href=\"/%s\" title=\"%s/%s -> %s\">%s/%s</a> ", base, Host, base, repo.Http, Host, base)
}

func main() {
	readRepos()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<html><head>")
			fmt.Fprintf(w, "<title>%s</title>", Title)
			fmt.Fprintf(w, "%s", Head)
			fmt.Fprintf(w, "</head><body>")
			fmt.Fprintf(w, "<pre>%s\n\n", Title)
			fmt.Fprintf(w, "Packages:\n")
			for basename, repo := range Repos {
				fmt.Fprintf(w, "&bullet; %s\n", printRepo(basename, repo))
			}
			fmt.Fprintf(w, "</pre>")
			fmt.Fprintf(w, "</body></html>")

			return
		}

		for basename, repo := range Repos {
			if req.URL.Path == "/"+basename {
				// if go-get header is not present redirect to http
				if !isGoGet(req) {
					http.Redirect(w, req, repo.Http, http.StatusPermanentRedirect)
					return
				}

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				fmt.Fprintf(w, "<html><head>")
				fmt.Fprintf(w, "<title>%s</title>", basename)
				fmt.Fprintf(w, "<meta name=\"go-import\" content=\"%s/%s git %s\">", Host, basename, repo.Git)
				// fmt.Fprintf(w, "<meta name=\"go-source\" content=\"%s/%s _ %s/tree{/dir} %s/tree{/dir}/{file}#n{line}\">", Host, basename, url, url)
				fmt.Fprintf(w, "</head>")
				fmt.Fprintf(w, "<body>go get %s/%s</body>", Host, basename)
				fmt.Fprintf(w, "</html>")
				return
			}
		}
	})
	http.ListenAndServe(ListenAddr, mux)
}
