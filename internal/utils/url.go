package utils

import (
	"hash/fnv"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

func HashUrl(url string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(url))
	return h.Sum64()
}

func GetHref(t html.Token, baseURL string) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			if len(a.Val) == 0 {
				return false, ""
			}

			// Skip fragments (same page anchors)
			if strings.HasPrefix(a.Val, "#") {
				return false, ""
			}

			// Skip javascript: and mailto: links
			if strings.HasPrefix(a.Val, "javascript:") || strings.HasPrefix(a.Val, "mailto:") {
				return false, ""
			}

			// Parse the base URL
			base, err := url.Parse(baseURL)
			if err != nil {
				return false, ""
			}

			// Parse the href (could be relative or absolute)
			ref, err := url.Parse(a.Val)
			if err != nil {
				return false, ""
			}

			// Resolve relative URL against base URL
			resolved := base.ResolveReference(ref)

			// Only crawl HTTP/HTTPS URLs
			if resolved.Scheme != "http" && resolved.Scheme != "https" {
				return false, ""
			}

			return true, resolved.String()
		}
	}
	return false, ""
}
