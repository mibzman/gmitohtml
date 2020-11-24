package gmitohtml

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var lastRequestTime = time.Now().Unix()

var clientCerts = make(map[string]tls.Certificate)

// ErrInvalidCertificate is the error returned when an invalid certificate is provided.
var ErrInvalidCertificate = errors.New("invalid certificate")

// Fetch downloads and converts a Gemini page.
func fetch(u string) ([]byte, []byte, error) {
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

	certHost := requestURL.Hostname()
	if strings.HasPrefix(certHost, "www.") {
		certHost = certHost[4:]
	}

	clientCert, certAvailable := clientCerts[certHost]
	if certAvailable {
		tlsConfig.Certificates = []tls.Certificate{clientCert}
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

	requestInput := bytes.HasPrefix(header, []byte("1"))
	if requestInput {
		requestSensitiveInput := bytes.HasPrefix(header, []byte("11"))

		data = []byte(inputPage)

		data = bytes.Replace(data, []byte("GEMINICURRENTURL"), []byte(rewriteURL(u, requestURL)), 1)

		currentURL := u
		if strings.HasPrefix(currentURL, "gemini://") {
			currentURL = currentURL[9:]
		}
		data = bytes.Replace(data, []byte("GEMINICURRENTURL"), []byte(currentURL), 1)

		prompt := "(No input prompt)"
		if len(header) > 3 {
			prompt = string(header[3:])
		}
		data = bytes.Replace(data, []byte("GEMINIINPUTPROMPT"), []byte(prompt), 1)

		inputType := "text"
		if requestSensitiveInput {
			inputType = "password"
		}
		data = bytes.Replace(data, []byte("GEMINIINPUTTYPE"), []byte(inputType), 1)

		return header, data, nil
	}

	if !bytes.HasPrefix(header, []byte("2")) {
		return header, []byte(fmt.Sprintf(pageHeader+"Server sent unexpected header:<br><br><b>%s</b>", header) + pageFooter), nil
	}

	if bytes.HasPrefix(header, []byte("20 text/html")) {
		return header, data, nil
	}
	return header, Convert(data, requestURL.String()), nil
}

func handleIndex(writer http.ResponseWriter, request *http.Request) {
	address := request.FormValue("address")
	if address != "" {
		http.Redirect(writer, request, rewriteURL(address, request.URL), http.StatusSeeOther)
		return
	}

	writer.Write([]byte(indexPage))
}

func handleRequest(writer http.ResponseWriter, request *http.Request) {
	defer request.Body.Close()

	lastRequestTime = time.Now().Unix()

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
	if request.URL.RawQuery != "" {
		u.RawQuery = request.URL.RawQuery
	}

	inputText := request.PostFormValue("input")
	if inputText != "" {
		u.RawQuery = inputText
		http.Redirect(writer, request, rewriteURL(u.String(), u), http.StatusSeeOther)
		return
	}

	header, data, err := fetch(u.String())
	if err != nil {
		fmt.Fprintf(writer, "Error: failed to fetch %s: %s", u, err)
		return
	}

	if len(header) > 0 && header[0] == '3' {
		split := bytes.SplitN(header, []byte(" "), 2)
		if len(split) == 2 {
			http.Redirect(writer, request, rewriteURL(string(split[1]), u), http.StatusSeeOther)
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

// LastRequestTime returns the time of the last request.
func LastRequestTime() int64 {
	return lastRequestTime
}

// SetClientCertificate sets the client certificate to use for a domain.
func SetClientCertificate(domain string, certificate []byte, privateKey []byte) error {
	if len(certificate) == 0 || len(privateKey) == 0 {
		delete(clientCerts, domain)
		return nil
	}

	clientCert, err := tls.X509KeyPair(certificate, privateKey)
	if err != nil {
		return ErrInvalidCertificate
	}

	leafCert, err := x509.ParseCertificate(clientCert.Certificate[0])
	if err == nil {
		clientCert.Leaf = leafCert
	}

	clientCerts[domain] = clientCert
	return nil
}
