package main

import (
	"context"
	"errors"
	"flag"
	"fmt"

	"github.com/genuinetools/reg/clair"
	"github.com/genuinetools/reg/registry"
	"github.com/sirupsen/logrus"
)

const vulnsHelp = `Get a vulnerability report for a repository from a CoreOS Clair server.`

func (cmd *vulnsCommand) Name() string      { return "vulns" }
func (cmd *vulnsCommand) Args() string      { return "[OPTIONS] NAME[:TAG|@DIGEST]" }
func (cmd *vulnsCommand) ShortHelp() string { return vulnsHelp }
func (cmd *vulnsCommand) LongHelp() string  { return vulnsHelp }
func (cmd *vulnsCommand) Hidden() bool      { return false }

func (cmd *vulnsCommand) Register(fs *flag.FlagSet) {
	fs.StringVar(&cmd.clairServer, "clair", "", "url to clair instance")
	fs.IntVar(&cmd.fixableThreshold, "fixable-threshhold", 0, "number of fixable issues permitted")
}

type vulnsCommand struct {
	clairServer      string
	fixableThreshold int
}

func (cmd *vulnsCommand) Run(ctx context.Context, args []string) error {
	if len(cmd.clairServer) < 1 {
		return errors.New("clair url cannot be empty, pass --clair")
	}

	if cmd.fixableThreshold < 0 {
		return errors.New("fixable threshold must be a positive integer")
	}

	if len(args) < 1 {
		return fmt.Errorf("pass the name of the repository")
	}

	image, err := registry.ParseImage(args[0])
	if err != nil {
		return err
	}

	// Create the registry client.
	r, err := createRegistryClient(image.Domain)
	if err != nil {
		return err
	}

	// Initialize clair client.
	cr, err := clair.New(cmd.clairServer, clair.Opt{
		Debug:    debug,
		Timeout:  timeout,
		Insecure: insecure,
	})
	if err != nil {
		return fmt.Errorf("creation of clair client at %s failed: %v", cmd.clairServer, err)
	}

	// Get the vulnerability report.
	report, err := cr.VulnerabilitiesV3(r, image.Path, image.Reference())
	if err != nil {
		// Fallback to Clair v2 API.
		report, err = cr.Vulnerabilities(r, image.Path, image.Reference())
		if err != nil {
			return err
		}
	}

	// Iterate over the vulnerabilities by severity list.
	for sev, vulns := range report.VulnsBySeverity {
		for _, v := range vulns {
			if sev == "Fixable" {
				fmt.Printf("%s: [%s] \n%s\n%s\n", v.Name, v.Severity+" - Fixable", v.Description, v.Link)
				fmt.Printf("Fixed by: %s\n", v.FixedBy)
			} else {
				fmt.Printf("%s: [%s] \n%s\n%s\n", v.Name, v.Severity, v.Description, v.Link)
			}
			fmt.Println("-----------------------------------------")
		}
	}

	// Print summary and count.
	for sev, vulns := range report.VulnsBySeverity {
		fmt.Printf("%s: %d\n", sev, len(vulns))
	}

	// Return an error if there are more than 1 fixable vulns.
	fixable, ok := report.VulnsBySeverity["Fixable"]
	if ok {
		if len(fixable) > cmd.fixableThreshold {
			logrus.Fatalf("%d fixable vulnerabilities found", len(fixable))
		}
	}

	// Return an error if there are more than 10 bad vulns.
	badVulns := 0
	// Include any high vulns.
	if highVulns, ok := report.VulnsBySeverity["High"]; ok {
		badVulns += len(highVulns)
	}
	// Include any critical vulns.
	if criticalVulns, ok := report.VulnsBySeverity["Critical"]; ok {
		badVulns += len(criticalVulns)
	}
	// Include any defcon1 vulns.
	if defcon1Vulns, ok := report.VulnsBySeverity["Defcon1"]; ok {
		badVulns += len(defcon1Vulns)
	}
	if badVulns > 10 {
		logrus.Fatalf("%d bad vulnerabilities found", badVulns)
	}

	return nil
}
