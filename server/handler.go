package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"

	mfst "github.com/fixate/redirect-server/manifest"
)

type handler struct {
	sync.RWMutex
	Manifest *mfst.Manifest
}

func newHandler(manifest *mfst.Manifest) *handler {
	return &handler{Manifest: manifest}
}

func isHttps(r *http.Request) bool {
	return strings.HasPrefix(r.Proto, "HTTPS") || r.Header.Get("X-Forwarded-Proto") == "https"
}

func (inst *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inst.RLock()
	defer inst.RUnlock()

	io.Copy(ioutil.Discard, r.Body)
	defer r.Body.Close()

	if inst.Manifest.Options.EnforceHttps && !isHttps(r) {
		var urlCopy url.URL
		urlCopy = *r.URL
		urlCopy.Scheme = "https"
		urlCopy.Host = r.Host
		inst.handleRedirect(w, r, urlCopy.String())
		return
	}

	url, redirect, found, err := inst.findRedirect(r)
	if err != nil {
		log.Fatal(err)
	}

	if found {
		fmt.Printf(
			"Redirect '%s/%s' found for '%s/%s'. Redirect to %s.",
			redirect.Host,
			redirect.Path,
			r.Host,
			normalizePath(r.URL.Path),
			url,
		)
		inst.handleRedirect(w, r, url)
		return
	}

	fmt.Println("Redirect not found.")

	w.WriteHeader(http.StatusNotFound)
}

func isPathMatch(a, b string) bool {
	return normalizePath(a) == normalizePath(b)
}

func normalizePath(path string) string {
	return strings.Trim(path, "/")
}

func convertRegex(str string) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(str)
	regexpStr := strings.Replace(escaped, "\\*", "(\\w*)", -1)
	return regexp.Compile(regexpStr)
}

func substitute(str string, matches []string) string {
	exprMatcher := regexp.MustCompile("(\\$[0-9]+)")
	return exprMatcher.ReplaceAllStringFunc(str, func(match string) string {
		index, err := strconv.Atoi(match[1:])
		if err != nil {
			return match
		}
		return matches[index]
	})
}

func (inst *handler) findRedirect(r *http.Request) (url string, result *mfst.Redirect, found bool, err error) {
	err = nil
	host := r.Host
	path := normalizePath(r.URL.Path)
	fmt.Printf("Checking redirect for %s/%s\n", host, path)

	for _, redirect := range inst.Manifest.Redirects {
		// If host or path is blank they pass
		isHostMatch := len(redirect.Host) == 0
		isPathMatch := len(redirect.Path) == 0
		if len(redirect.Host) > 0 {
			hostRegex, rerr := convertRegex(redirect.Host)
			if rerr != nil {
				err = rerr
				return
			}
			isHostMatch = hostRegex.MatchString(host)
			url = redirect.Target
		}

		if len(redirect.Path) > 0 {
			pathRegex, rerr := convertRegex(redirect.Path)
			if rerr != nil {
				err = rerr
				return
			}
			pathMatches := pathRegex.FindStringSubmatch(path)
			if len(pathMatches) == 0 {
				isPathMatch = false
			} else {
				isPathMatch = true
				fmt.Println(redirect)
				url = substitute(redirect.Target, pathMatches)
			}
		}

		if isHostMatch && isPathMatch {
			result = &redirect
			found = true
			return
		}
	}

	found = false
	return
}

func (inst *handler) handleRedirect(w http.ResponseWriter, r *http.Request, url string) {
	if r.Method == "HEAD" {
		r.Close = true
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Add("Content-Length", "0")
	http.Redirect(w, r, url, http.StatusFound)
}
