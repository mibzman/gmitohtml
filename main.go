package main

import (
	"gitlab.com/tslocum/gmitohtml/pkg/gmitohtml"

	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
)

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var view bool
	var daemon string
	flag.BoolVar(&view, "view", false, "open web browser")
	flag.StringVar(&daemon, "daemon", "", "start daemon on specified address")
	// TODO option to include response header in page
	flag.Parse()

	if daemon != "" {
		err := gmitohtml.StartDaemon(daemon)
		if err != nil {
			log.Fatal(err)
		}

		if view {
			openBrowser("http://" + daemon)
		}

		select {} //TODO
	}

	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	data = gmitohtml.Convert(data, "")

	if view {
		openBrowser(string(append([]byte("data:text/html,"), []byte(url.PathEscape(string(data)))...)))
		return
	}
	fmt.Print(gmitohtml.Convert(data, ""))
}
