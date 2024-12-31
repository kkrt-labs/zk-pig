package url

import (
	"fmt"
	"net/url"
	"strings"
)

// Parse parses the base URL, performs some validations and returns a url.URL.
func Parse(baseURL string) (*url.URL, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		return nil, fmt.Errorf("no scheme detected in URL %s", baseURL)
	}
	if u.RawQuery != "" {
		// handle unencoded semicolons
		u.RawQuery = strings.Replace(u.RawQuery, ";", "%3B", -1)
		var q url.Values
		q, err = url.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, err
		}
		u.RawQuery = q.Encode()
	}
	return u, nil
}
