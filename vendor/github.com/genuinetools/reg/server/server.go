package main

import (
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/genuinetools/reg/clair"
	"github.com/genuinetools/reg/registry"
	"github.com/genuinetools/reg/repoutils"
	"github.com/gorilla/mux"
	wordwrap "github.com/mitchellh/go-wordwrap"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	// VERSION is the binary version.
	VERSION = "v0.2.0"
)

var (
	updating = false
	r        *registry.Registry
	cl       *clair.Clair
	tmpl     *template.Template
)

// preload initializes any global options and configuration
// before the main or sub commands are run.
func preload(c *cli.Context) (err error) {
	if c.GlobalBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "reg-server"
	app.Version = VERSION
	app.Author = "The Genuinetools Authors"
	app.Email = "no-reply@butts.com"
	app.Usage = "Docker registry v2 static UI server."
	app.Before = preload
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "run in debug mode",
		},
		cli.StringFlag{
			Name:  "username, u",
			Usage: "username for the registry",
		},
		cli.StringFlag{
			Name:  "password, p",
			Usage: "password for the registry",
		},
		cli.StringFlag{
			Name:  "registry, r",
			Usage: "URL to the private registry (ex. r.j3ss.co)",
		},
		cli.BoolFlag{
			Name:  "insecure, k",
			Usage: "do not verify tls certificates of registry",
		},
		cli.BoolFlag{
			Name:  "once, o",
			Usage: "generate an output once and then exit",
		},
		cli.StringFlag{
			Name:  "port",
			Value: "8080",
			Usage: "port for server to run on",
		},
		cli.StringFlag{
			Name:  "cert",
			Usage: "path to ssl cert",
		},
		cli.StringFlag{
			Name:  "key",
			Usage: "path to ssl key",
		},
		cli.StringFlag{
			Name:  "interval",
			Value: "1h",
			Usage: "interval to generate new index.html's at",
		},
		cli.StringFlag{
			Name:  "clair",
			Usage: "url to clair instance",
		},
	}
	app.Action = func(c *cli.Context) error {
		auth, err := repoutils.GetAuthConfig(c.GlobalString("username"), c.GlobalString("password"), c.GlobalString("registry"))
		if err != nil {
			logrus.Fatal(err)
		}

		// create the registry client
		if c.GlobalBool("insecure") {
			r, err = registry.NewInsecure(auth, c.GlobalBool("debug"))
			if err != nil {
				logrus.Fatal(err)
			}
		} else {
			r, err = registry.New(auth, c.GlobalBool("debug"))
			if err != nil {
				logrus.Fatal(err)
			}
		}

		// create a clair instance if needed
		if c.GlobalString("clair") != "" {
			cl, err = clair.New(c.GlobalString("clair"), c.GlobalBool("debug"))
			if err != nil {
				logrus.Warnf("creation of clair failed: %v", err)
			}
		}

		// get the path to the static directory
		wd, err := os.Getwd()
		if err != nil {
			logrus.Fatal(err)
		}
		staticDir := filepath.Join(wd, "static")

		// create the template
		templateDir := filepath.Join(staticDir, "../templates")

		// make sure all the templates exist
		vulns := filepath.Join(templateDir, "vulns.html")
		if _, err := os.Stat(vulns); os.IsNotExist(err) {
			logrus.Fatalf("Template %s not found", vulns)
		}
		layout := filepath.Join(templateDir, "repositories.html")
		if _, err := os.Stat(layout); os.IsNotExist(err) {
			logrus.Fatalf("Template %s not found", layout)
		}
		tags := filepath.Join(templateDir, "tags.html")
		if _, err := os.Stat(tags); os.IsNotExist(err) {
			logrus.Fatalf("Template %s not found", tags)
		}

		funcMap := template.FuncMap{
			"trim": func(s string) string {
				return wordwrap.WrapString(s, 80)
			},
			"color": func(s string) string {
				switch s = strings.ToLower(s); s {
				case "high":
					return "danger"
				case "critical":
					return "danger"
				case "defcon1":
					return "danger"
				case "medium":
					return "warning"
				case "low":
					return "info"
				case "negligible":
					return "info"
				case "unknown":
					return "default"
				default:
					return "default"
				}
			},
		}

		tmpl = template.Must(template.New("").Funcs(funcMap).ParseGlob(templateDir + "/*.html"))

		rc := registryController{
			reg: r,
			cl:  cl,
		}

		// create the initial index
		logrus.Info("creating initial static index")
		if err := rc.repositories(staticDir); err != nil {
			logrus.Fatalf("Error creating index: %v", err)
		}

		if c.GlobalBool("once") {
			logrus.Info("Output generated")
			return nil
		}

		// parse the duration
		dur, err := time.ParseDuration(c.String("interval"))
		if err != nil {
			logrus.Fatalf("parsing %s as duration failed: %v", c.String("interval"), err)
		}
		ticker := time.NewTicker(dur)

		go func() {
			// create more indexes every X minutes based off interval
			for range ticker.C {
				if !updating {
					logrus.Info("creating timer based static index")
					if err := rc.repositories(staticDir); err != nil {
						logrus.Warnf("creating static index failed: %v", err)
						updating = false
					}
				} else {
					logrus.Warnf("skipping timer based static index update for %s", c.String("interval"))
				}
			}
		}()

		// create mux server
		mux := mux.NewRouter()
		mux.UseEncodedPath()

		// static files handler
		staticHandler := http.FileServer(http.Dir(staticDir))
		mux.HandleFunc("/repo/{repo}/tags", rc.tagsHandler)
		mux.HandleFunc("/repo/{repo}/tags/", rc.tagsHandler)
		mux.HandleFunc("/repo/{repo}/tag/{tag}", rc.vulnerabilitiesHandler)
		mux.HandleFunc("/repo/{repo}/tag/{tag}/", rc.vulnerabilitiesHandler)
		mux.HandleFunc("/repo/{repo}/tag/{tag}/vulns", rc.vulnerabilitiesHandler)
		mux.HandleFunc("/repo/{repo}/tag/{tag}/vulns/", rc.vulnerabilitiesHandler)
		mux.HandleFunc("/repo/{repo}/tag/{tag}/vulns.json", rc.vulnerabilitiesHandler)
		mux.PathPrefix("/static/").Handler(http.StripPrefix("/static/", staticHandler))
		mux.Handle("/", staticHandler)

		// set up the server
		port := c.String("port")
		server := &http.Server{
			Addr:    ":" + port,
			Handler: mux,
		}
		logrus.Infof("Starting server on port %q", port)
		if c.String("cert") != "" && c.String("key") != "" {
			logrus.Fatal(server.ListenAndServeTLS(c.String("cert"), c.String("key")))
		} else {
			logrus.Fatal(server.ListenAndServe())
		}

		return nil
	}

	app.Run(os.Args)
}
