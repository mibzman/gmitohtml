package gmitohtml

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html"
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
		scheme := "gemini"
		if strings.HasPrefix(loc.Path, "/file/") {
			scheme = "file"
		}

		if strings.HasPrefix(u, "file://") {
			if !allowFileAccess {
				return "http://" + daemonAddress + "/?FileAccessNotAllowed"
			}
			return "http://" + daemonAddress + "/file/" + u[7:]
		}

		offset := 0
		if strings.HasPrefix(u, "gemini://") {
			offset = 9
		}
		firstSlash := strings.IndexRune(u[offset:], '/')
		if firstSlash != -1 {
			u = strings.ToLower(u[:firstSlash+offset]) + u[firstSlash+offset:]
		}

		if strings.HasPrefix(u, "gemini://") {
			return "http://" + daemonAddress + "/gemini/" + u[9:]
		} else if strings.Contains(u, "://") {
			return u
		} else if loc != nil && len(u) > 0 && !strings.HasPrefix(u, "//") {
			if u[0] != '/' {
				if loc.Path[len(loc.Path)-1] == '/' {
					u = path.Join("/", loc.Path, u)
				} else {
					u = path.Join("/", path.Dir(loc.Path), u)
				}
			}
			return "http://" + daemonAddress + "/" + scheme + "/" + strings.ToLower(loc.Host) + u
		}
		return "http://" + daemonAddress + "/" + scheme + "/" + u
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
			result = append(result, html.EscapeString(string(line))...)
			result = append(result, []byte("\n")...)
			continue
		}

		if l >= 6 && bytes.HasPrefix(line, []byte("=>")) {
			splitStart := 2
			if line[splitStart] == ' ' || line[splitStart] == '\t' {
				splitStart++
			}

			var split [][]byte
			firstSpace := bytes.IndexRune(line[splitStart:], ' ')
			firstTab := bytes.IndexRune(line[splitStart:], '\t')
			if firstSpace != -1 && (firstTab == -1 || firstSpace < firstTab) {
				split = bytes.SplitN(line[splitStart:], []byte(" "), 2)
			} else if firstTab != -1 {
				split = bytes.SplitN(line[splitStart:], []byte("\t"), 2)
			}

			var linkURL []byte
			var linkLabel []byte
			if len(split) == 2 {
				linkURL = split[0]
				linkLabel = split[1]
			} else {
				linkURL = line[splitStart:]
				linkLabel = line[splitStart:]
			}

			link := append([]byte(`<a href="`), html.EscapeString(rewriteURL(string(linkURL), parsedURL))...)
			link = append(link, []byte(`">`)...)
			link = append(link, html.EscapeString(string(linkLabel))...)
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
			result = append(result, []byte(fmt.Sprintf("<h%d>%s</h%d>", heading, html.EscapeString(string(line[heading:])), heading))...)
			continue
		}

		result = append(result, html.EscapeString(string(line))...)
		result = append(result, []byte("<br>")...)
	}

	if preformatted {
		result = append(result, []byte("</pre>\n")...)
	}

	result = append([]byte(pageHeader), result...)
	result = append(result, []byte(pageFooter)...)
	return fillTemplateVariables(result, u, false)
}
