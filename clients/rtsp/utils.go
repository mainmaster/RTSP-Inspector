package rtsp

import (
	"regexp"
)

const (
	realmRegEx     = `realm="([^"]+)"`
	nonceRegEx     = `nonce="([^"]+)"`
	sessionIDRegEx = `Session: ([^; \r\n]+)`
	authHeader     = "WWW-Authenticate"
)

func findParam(header, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(header)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
