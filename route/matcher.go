package route

import (
	"strings"
)

// matcher determines whether a host/path matches a route
type urlMatcher func(uri string, r *Route) bool
type pathStripper func(path string, t *Target) string

type matcher struct {
	UrlMatcher urlMatcher
	PathStripper pathStripper
}

// Matcher contains the available matcher functions.
// Update config/load.go#load after updating.
var Matcher = map[string]matcher{
	"prefix":  {prefixMatcher, prefixPathStripper},
	"glob":    {globMatcher, globPathStripper},
	"iprefix": {iPrefixMatcher, iPrefixPathStripper},
	"regex": {regexMatcher, regexPathStripper},
}

// prefixMatcher matches path to the routes' path.
func prefixMatcher(uri string, r *Route) bool {
	return strings.HasPrefix(uri, r.Path)
}

func prefixPathStripper(path string, t *Target) string {
	if strings.HasPrefix(path, t.StripPath) {
		return path[len(t.StripPath):]
	}  else {
		return path
	}
}

// globMatcher matches path to the routes' path using gobwas/glob.
func globMatcher(uri string, r *Route) bool {
	return r.Glob.Match(uri)
}

func globPathStripper(path string, t *Target) string {
	return iPrefixPathStripper(path, t)
}

// iPrefixMatcher matches path to the routes' path ignoring case
func iPrefixMatcher(uri string, r *Route) bool {
	// todo(fs): if this turns out to be a performance issue we should cache
	// todo(fs): strings.ToLower(r.Path) in r.PathLower
	lowerURI := strings.ToLower(uri)
	lowerPath := strings.ToLower(r.Path)
	return strings.HasPrefix(lowerURI, lowerPath)
}

func iPrefixPathStripper(path string, t *Target) string {
	lowerPath := strings.ToLower(path)
	lowerPathStrip := strings.ToLower(t.StripPath)

	if strings.HasPrefix(lowerPath, lowerPathStrip) {
		return path[len(lowerPathStrip):]
	}  else {
		return path
	}
}

func regexMatcher(uri string, r *Route) bool {
	return r.Regex.MatchString(uri)
}

func regexPathStripper(path string, t *Target) string {
	tmp := path
	if t.StripPathRegex != nil {
		for _, item := range t.StripPathRegex.FindAllString(path, -1) {
			tmp = strings.ReplaceAll(tmp, item, "")
		}
	}

	return tmp
}
