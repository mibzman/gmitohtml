package gmitohtml

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// ErrInvalidURL is the error returned when the URL is invalid.
var ErrInvalidURL = errors.New("invalid URL")

var daemonAddress string

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

	var lastPreformatted bool
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
			continue
		}

		if preformatted != lastPreformatted {
			if preformatted {
				result = append(result, []byte("<pre>\n")...)
			} else {
				result = append(result, []byte("</pre>\n")...)
			}
			lastPreformatted = preformatted
		}

		if preformatted {
			result = append(result, line...)
			result = append(result, []byte("\n")...)
			continue
		}

		if l >= 7 && bytes.HasPrefix(line, []byte("=> ")) {
			split := bytes.SplitN(line[3:], []byte(" "), 2)
			if len(split) == 2 {
				link := append([]byte(`<a href="`), rewriteURL(string(split[0]), parsedURL)...)
				link = append(link, []byte(`">`)...)
				link = append(link, split[1]...)
				link = append(link, []byte(`</a>`)...)
				result = append(result, link...)
				result = append(result, []byte("<br>")...)
				continue
			}
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
			result = append(result, []byte(fmt.Sprintf("<h%d>%s</h%d>", heading, line, heading))...)
			continue
		}

		result = append(result, line...)
		result = append(result, []byte("<br>")...)
	}

	return result
}

// Fetch downloads and converts a Gemini page.
func Fetch(u string, clientCertFile string, clientCertKey string) ([]byte, []byte, error) {
	if u == "" {
		return nil, nil, ErrInvalidURL
	}

	requestURL, err := url.ParseRequestURI(u)
	if err != nil {
		return nil, nil, err
	}
	if requestURL.Scheme == "" {
		requestURL.Scheme = "gemini"
	}
	if strings.IndexRune(requestURL.Host, ':') == -1 {
		requestURL.Host += ":1965"
	}

	tlsConfig := &tls.Config{}

	conn, err := tls.Dial("tcp", requestURL.Host, tlsConfig)
	if err != nil {
		return nil, nil, err
	}

	// Send request header
	conn.Write([]byte(requestURL.String() + "\r\n"))

	data, err := ioutil.ReadAll(conn)
	if err != nil {
		return nil, nil, err
	}

	firstNewLine := -1
	l := len(data)
	if l > 2 {
		for i := 1; i < l; i++ {
			if data[i] == '\n' && data[i-1] == '\r' {
				firstNewLine = i
				break
			}
		}
	}
	var header []byte
	if firstNewLine > -1 {
		header = data[:firstNewLine]
		data = data[firstNewLine+1:]
	}

	if bytes.HasPrefix(header, []byte("text/html")) {
		return header, data, nil
	}
	return header, Convert(data, requestURL.String()), nil
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	if request.URL == nil {
		return
	}

	pathSplit := strings.Split(request.URL.Path, "/")
	if len(pathSplit) < 2 || pathSplit[1] != "gemini" {
		writer.Write([]byte("Error: invalid protocol, only Gemini is supported"))
		return
	}

	u, err := url.ParseRequestURI("gemini://" + strings.Join(pathSplit[2:], "/"))
	if err != nil {
		writer.Write([]byte("Error: invalid URL"))
		return
	}

	header, data, err := Fetch(u.String(), "", "")
	if err != nil {
		fmt.Fprintf(writer, "Error: failed to fetch %s: %s", u, err)
		return
	}

	if len(header) > 0 && header[0] == '3' {
		split := bytes.SplitN(header, []byte(" "), 2)
		if len(split) == 2 {
			writer.Header().Set("Location", rewriteURL(string(split[1]), request.URL))
			return
		}
	}

	if len(header) > 3 && !bytes.HasPrefix(header[3:], []byte("text/gemini")) {
		writer.Header().Set("Content-type", string(header[3:]))
	} else {
		writer.Header().Set("Content-type", "text/html; charset=utf-8")
	}

	writer.Write([]byte("<!DOCTYPE html>\n<html>\n<body>\n"))
	writer.Write(data)
	writer.Write([]byte("\n</body>\n</html>"))
}

// StartDaemon starts the page conversion daemon.
func StartDaemon(address string) error {
	daemonAddress = address

	http.HandleFunc("/", handleRequest)

	go func() {
		log.Fatal(http.ListenAndServe(address, nil))
	}()
	return nil
}
