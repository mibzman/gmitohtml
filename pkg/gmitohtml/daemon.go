package gmitohtml

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Fetch downloads and converts a Gemini page.
func fetch(u string, clientCertFile string, clientCertKey string) ([]byte, []byte, error) {
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

	host := requestURL.Host
	if strings.IndexRune(host, ':') == -1 {
		host += ":1965"
	}

	tlsConfig := &tls.Config{
		// This must be enabled until most sites have transitioned away from
		// using self-signed certificates.
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", host, tlsConfig)
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

func handleIndex(writer http.ResponseWriter, request *http.Request) {
	address := request.FormValue("address")
	if address != "" {
		http.Redirect(writer, request, rewriteURL(address, request.URL), http.StatusTemporaryRedirect)
		return
	}

	writer.Write([]byte(indexPage))
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()
	if request.URL == nil {
		return
	}

	if request.URL.Path == "/" {
		handleIndex(writer, request)
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

	header, data, err := fetch(u.String(), "", "")
	if err != nil {
		fmt.Fprintf(writer, "Error: failed to fetch %s: %s", u, err)
		return
	}

	if len(header) > 0 && header[0] == '3' {
		split := bytes.SplitN(header, []byte(" "), 2)
		if len(split) == 2 {
			http.Redirect(writer, request, rewriteURL(string(split[1]), request.URL), http.StatusTemporaryRedirect)
			return
		}
	}

	if len(header) > 3 && !bytes.HasPrefix(header[3:], []byte("text/gemini")) {
		writer.Header().Set("Content-Type", string(header[3:]))
	} else {
		writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	writer.Write(data)
}

func handleAssets(writer http.ResponseWriter, request *http.Request) {
	assetLock.Lock()
	defer assetLock.Unlock()

	writer.Header().Set("Cache-Control", "max-age=86400")

	http.FileServer(fs).ServeHTTP(writer, request)
}

// StartDaemon starts the page conversion daemon.
func StartDaemon(address string) error {
	loadAssets()

	daemonAddress = address

	handler := http.NewServeMux()
	handler.HandleFunc("/assets/style.css", handleAssets)
	handler.HandleFunc("/", handleRequest)
	go func() {
		log.Fatal(http.ListenAndServe(address, handler))
	}()

	return nil
}
