package gmitohtml

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"sync"
)

// ErrInvalidURL is the error returned when the URL is invalid.
var ErrInvalidURL = errors.New("invalid URL")

var daemonAddress string

var assetLock sync.Mutex

func rewriteURL(u string, loc *url.URL) string {
	if daemonAddress != "" {
		if strings.HasPrefix(u, "gemini://") {
			return "http://" + daemonAddress + "/gemini/" + u[9:]
		} else if strings.Contains(u, "://") {
			return u
		} else if loc != nil && len(u) > 0 && !strings.HasPrefix(u, "//") {
			newPath := u
			if u[0] != '/' {
				newPath = path.Join(loc.Path, u)
			}
			return "http://" + daemonAddress + "/gemini/" + loc.Host + newPath
		}
		return "http://" + daemonAddress + "/gemini/" + u
	}
	return u
}

// Convert converts text/gemini to text/html.
func Convert(page []byte, u string) []byte {
	var result []byte

	var preformatted bool

	parsedURL, err := url.Parse(u)
	if err != nil {
		parsedURL = nil
		err = nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(page))
	for scanner.Scan() {
		line := scanner.Bytes()
		l := len(line)
		if l >= 3 && string(line[0:3]) == "```" {
			preformatted = !preformatted
			if preformatted {
				result = append(result, []byte("<pre>\n")...)
			} else {
				result = append(result, []byte("</pre>\n")...)
			}
			continue
		}

		if preformatted {
			result = append(result, line...)
			result = append(result, []byte("\n")...)
			continue
		}

		if l >= 6 && bytes.HasPrefix(line, []byte("=>")) {
			splitStart := 2
			if line[splitStart] == ' ' || line[splitStart] == '\t' {
				splitStart++
			}
			split := bytes.SplitN(line[splitStart:], []byte(" "), 2)
			if len(split) != 2 {
				split = bytes.SplitN(line[splitStart:], []byte("\t"), 2)
			}

			linkURL := line[splitStart:]
			linkLabel := line[splitStart:]
			if len(split) == 2 {
				linkURL = split[0]
				linkLabel = split[1]
			}
			link := append([]byte(`<a href="`), rewriteURL(string(linkURL), parsedURL)...)
			link = append(link, []byte(`">`)...)
			link = append(link, linkLabel...)
			link = append(link, []byte(`</a>`)...)
			result = append(result, link...)
			result = append(result, []byte("<br>")...)
			continue
		}

		heading := 0
		for i := 0; i < l; i++ {
			if line[i] == '#' {
				heading++
			} else {
				break
			}
		}
		if heading > 0 {
			result = append(result, []byte(fmt.Sprintf("<h%d>%s</h%d>", heading, line[heading:], heading))...)
			continue
		}

		result = append(result, line...)
		result = append(result, []byte("<br>")...)
	}

	if preformatted {
		result = append(result, []byte("</pre>\n")...)
	}

	result = append([]byte(pageHeader), result...)
	result = append(result, []byte(pageFooter)...)
	return fillTemplateVariables(result, u, false)
}
