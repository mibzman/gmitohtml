package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"

	"github.com/mibzman/gmitohtml/pkg/gmitohtml"
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
	var (
		view       bool
		allowFile  bool
		daemon     string
		hostname   string
		configFile string
	)
	flag.BoolVar(&view, "view", false, "open web browser")
	flag.BoolVar(&allowFile, "allow-file", false, "allow local file access via file://")
	flag.StringVar(&daemon, "daemon", "", "start daemon on specified address")
	flag.StringVar(&hostname, "hostname", "", "server hostname (e.g. rocketnine.space) (defaults to daemon address)")
	flag.StringVar(&configFile, "config", "", "path to configuration file")
	// TODO option to include response header in page
	flag.Parse()

	defaultConfig := defaultConfigPath()
	if configFile == "" {
		configFile = defaultConfig
	}

	if configFile != "" {
		var configExists bool
		if _, err := os.Stat(defaultConfig); !os.IsNotExist(err) {
			configExists = true
		}

		if configExists || configFile != defaultConfig {
			err := readconfig(configFile)
			if err != nil {
				log.Fatalf("failed to read configuration file at %s: %v\nSee CONFIGURATION.md for information on configuring gmitohtml", configFile, err)
			}

			for u, label := range config.Bookmarks {
				gmitohtml.AddBookmark(u, label)
			}
		}
	}

	for domain, cc := range config.Certs {
		certData, err := ioutil.ReadFile(cc.Cert)
		if err != nil {
			log.Fatalf("failed to load client certificate for domain %s: %s", domain, err)
		}

		keyData, err := ioutil.ReadFile(cc.Key)
		if err != nil {
			log.Fatalf("failed to load client certificate for domain %s: %s", domain, err)
		}

		err = gmitohtml.SetClientCertificate(domain, certData, keyData)
		if err != nil {
			log.Fatalf("failed to load client certificate for domain %s", domain)
		}
	}

	if daemon != "" {
		gmitohtml.SetOnBookmarksChanged(func() {
			config.Bookmarks = gmitohtml.GetBookmarks()

			err := saveConfig(configFile)
			if err != nil {
				log.Fatal(err)
			}
		})

		err := gmitohtml.StartDaemon(daemon, hostname, allowFile)
		if err != nil {
			log.Fatal(err)
		}

		if view {
			openBrowser("http://" + daemon)
		}

		select {}
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
