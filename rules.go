package cache

import (
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol/cacheobject"
)

// CacheRule determines if a request should be cached
type CacheRule interface {
	matches(*http.Request, int, *http.Header) bool
}

// PathCacheRule matches if the request starts with given Path
type PathCacheRule struct {
	Path string
}

// HeaderCacheRule matches if given Header matches any of the values
type HeaderCacheRule struct {
	Header string
	Value  []string
}

/* This rules decide if the request must be cached and are added to handler config if are present in Caddyfile */

func (rule *PathCacheRule) matches(req *http.Request, statusCode int, respHeaders *http.Header) bool {
	return strings.HasPrefix(req.URL.Path, rule.Path)
}

func (rule *HeaderCacheRule) matches(req *http.Request, statusCode int, respHeaders *http.Header) bool {
	headerValue := respHeaders.Get(rule.Header)
	for _, expectedValue := range rule.Value {
		if expectedValue == headerValue {
			return true
		}
	}
	return false
}

func getCacheableStatus(req *http.Request, response *Response, config *Config) (bool, time.Time, error) {
	reasonsNotToCache, expiration, err := cacheobject.UsingRequestResponse(req, response.Code, response.HeaderMap, false)

	if err != nil {
		return false, time.Now(), err
	}

	canBeStored := len(reasonsNotToCache) == 0

	if !canBeStored {
		return false, time.Now(), nil
	}

	varyHeader := response.HeaderMap.Get("Vary")
	if varyHeader == "*" {
		return false, time.Now(), nil
	}

	return expiration.After(time.Now().UTC()), expiration, nil
}

func matchesVary(currentRequest *http.Request, previousResponse *Response) bool {
	vary := previousResponse.HeaderMap.Get("Vary")

	for _, searchedHeader := range strings.Split(vary, ",") {
		searchedHeader = strings.TrimSpace(searchedHeader)
		if !reflect.DeepEqual(currentRequest.Header[searchedHeader], previousResponse.HeaderMap[searchedHeader]) {
			return false
		}
	}

	return true
}
