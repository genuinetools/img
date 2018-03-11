package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/genuinetools/reg/clair"
	"github.com/genuinetools/reg/registry"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type registryController struct {
	reg *registry.Registry
	cl  *clair.Clair
}

type v1Compatibility struct {
	ID      string    `json:"id"`
	Created time.Time `json:"created"`
}

// A Repository holds data after a vulnerability scan of a single repo
type Repository struct {
	Name                string                    `json:"name"`
	Tag                 string                    `json:"tag"`
	Created             time.Time                 `json:"created"`
	URI                 string                    `json:"uri"`
	VulnerabilityReport clair.VulnerabilityReport `json:"vulnerability"`
}

// A AnalysisResult holds all vulnerabilities of a scan
type AnalysisResult struct {
	Repositories   []Repository `json:"repositories"`
	RegistryDomain string       `json:"registryDomain"`
	Name           string       `json:"name"`
	LastUpdated    string       `json:"lastUpdated"`
}

func (rc *registryController) repositories(staticDir string) error {
	updating = true
	logrus.Info("fetching catalog")

	result := AnalysisResult{
		RegistryDomain: rc.reg.Domain,
		LastUpdated:    time.Now().Local().Format(time.RFC1123),
	}

	repoList, err := r.Catalog("")
	if err != nil {
		return fmt.Errorf("getting catalog failed: %v", err)
	}

	for _, repo := range repoList {
		repoURI := fmt.Sprintf("%s/%s", rc.reg.Domain, repo)
		r := Repository{
			Name: repo,
			URI:  repoURI,
		}

		result.Repositories = append(result.Repositories, r)
	}

	// parse & execute the template
	logrus.Info("executing the template repositories")

	path := filepath.Join(staticDir, "index.html")
	if err := os.MkdirAll(filepath.Dir(path), 0644); err != nil {
		return err
	}
	logrus.Debugf("creating/opening file %s", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.ExecuteTemplate(f, "repositories", result); err != nil {
		f.Close()
		return fmt.Errorf("execute template repositories failed: %v", err)
	}

	updating = false
	return nil
}

func (rc *registryController) tagsHandler(w http.ResponseWriter, r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"func":   "tags",
		"URL":    r.URL,
		"method": r.Method,
	}).Info("fetching tags")

	vars := mux.Vars(r)
	repo, err := url.QueryUnescape(vars["repo"])
	if err != nil || repo == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Empty repo")
		return
	}

	tags, err := rc.reg.Tags(repo)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"func":   "tags",
			"URL":    r.URL,
			"method": r.Method,
		}).Errorf("getting tags for %s failed: %v", repo, err)

		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "No tags found")
		return
	}

	// Error out if there are no tags / images (the above err != nil does not error out when nothing has been found)
	if len(tags) == 0 {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "No tags found")
		return
	}

	result := AnalysisResult{
		RegistryDomain: rc.reg.Domain,
		LastUpdated:    time.Now().Local().Format(time.RFC1123),
		Name:           repo,
	}

	for _, tag := range tags {
		// get the manifest
		m1, err := rc.reg.ManifestV1(repo, tag)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"func":   "tags",
				"URL":    r.URL,
				"method": r.Method,
				"repo":   repo,
				"tag":    tag,
			}).Errorf("getting v1 manifest for %s:%s failed: %v", repo, tag, err)
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, "Manifest not found")
			return
		}

		var createdDate time.Time
		for _, h := range m1.History {
			var comp v1Compatibility

			if err := json.Unmarshal([]byte(h.V1Compatibility), &comp); err != nil {
				logrus.WithFields(logrus.Fields{
					"func":   "tags",
					"URL":    r.URL,
					"method": r.Method,
				}).Errorf("unmarshal v1 manifest for %s:%s failed: %v", repo, tag, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			createdDate = comp.Created
			break
		}

		repoURI := fmt.Sprintf("%s/%s", rc.reg.Domain, repo)
		if tag != "latest" {
			repoURI += ":" + tag
		}
		rp := Repository{
			Name:    repo,
			Tag:     tag,
			URI:     repoURI,
			Created: createdDate,
		}

		result.Repositories = append(result.Repositories, rp)
	}

	if err := tmpl.ExecuteTemplate(w, "tags", result); err != nil {
		logrus.WithFields(logrus.Fields{
			"func":   "tags",
			"URL":    r.URL,
			"method": r.Method,
		}).Errorf("template rendering failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (rc *registryController) vulnerabilitiesHandler(w http.ResponseWriter, r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"func":   "vulnerabilities",
		"URL":    r.URL,
		"method": r.Method,
	}).Info("fetching vulnerabilities")

	vars := mux.Vars(r)
	repo, err := url.QueryUnescape(vars["repo"])
	tag := vars["tag"]

	if err != nil || repo == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Empty repo")
		return
	}

	if tag == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "Empty tag")
		return
	}

	result := clair.VulnerabilityReport{}

	if rc.cl != nil {
		result, err = rc.cl.Vulnerabilities(rc.reg, repo, tag)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"func":   "vulnerabilities",
				"URL":    r.URL,
				"method": r.Method,
			}).Errorf("vulnerability scanning for %s:%s failed: %v", repo, tag, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	if strings.HasSuffix(r.URL.String(), ".json") {
		js, err := json.Marshal(result)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"func":   "vulnerabilities",
				"URL":    r.URL,
				"method": r.Method,
			}).Errorf("json marshal failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "vulns", result); err != nil {
		logrus.WithFields(logrus.Fields{
			"func":   "vulnerabilities",
			"URL":    r.URL,
			"method": r.Method,
		}).Errorf("template rendering failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
