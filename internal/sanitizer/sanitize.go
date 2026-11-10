package sanitizer

import (
	"regexp"
)

var reAPIkey = regexp.MustCompile(`sk-[a-zA-Z0-9_-]{20,}`)
var reEmail = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
var rePhoneNum = regexp.MustCompile(`\+?\d[\d\s\-\(\)]{9,}\d`)

func Clean(body []byte) []byte {
	body = reAPIkey.ReplaceAll(body, []byte("[REDACTED]"))
	body = reEmail.ReplaceAll(body, []byte("[REDACTED]"))
	body = rePhoneNum.ReplaceAll(body, []byte("[REDACTED]"))
	return body
}
