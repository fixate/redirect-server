package server

import (
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

func pathMatchesHealthCheck(r *http.Request, path string) bool {
	return len(path) > 0 && normalizePath(r.URL.Path) == normalizePath(path)
}

func (inst *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	inst.RLock()
	defer inst.RUnlock()

	io.Copy(ioutil.Discard, r.Body)
	defer r.Body.Close()

	healthCheckPath := inst.Manifest.Options.HealthCheck
	if pathMatchesHealthCheck(r, healthCheckPath) {
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Length", "0")
		return
	}

	if inst.Manifest.Options.EnforceHttps && !isHttps(r) {
		var urlCopy url.URL
		urlCopy = *r.URL
		urlCopy.Scheme = "https"
		urlCopy.Host = r.Host
		inst.handleRedirect(w, r, urlCopy.String())
		return
	}

	url, _, found, err := inst.findRedirect(r)
	if err != nil {
		log.Fatal(err)
	}

	if found {
		inst.handleRedirect(w, r, url)
		return
	}

	log.Println("Redirect not found.")

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
	log.Printf("Checking redirect for %s/%s\n", host, path)

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
				url = substitute(redirect.Target, pathMatches)
			}
		}

		if len(redirect.PathMatch) > 0 {
			isPathMatch = false
			re, rerr := regexp.Compile(redirect.PathMatch)
			if rerr != nil {
				log.Println(rerr)
				err = rerr
				return
			}
			if re.MatchString(path) {
				isPathMatch = true
				url = re.ReplaceAllString(path, redirect.Target)
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
	log.Printf(
		"Redirect '%s/%s' -> '%s'",
		r.Host,
		normalizePath(r.URL.Path),
		url,
	)

	if r.Method == "HEAD" {
		r.Close = true
		w.Header().Add("Location", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Add("Content-Length", "0")
	http.Redirect(w, r, url, http.StatusFound)
}
