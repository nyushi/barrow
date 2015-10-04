package main

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/alecthomas/kingpin"
	"github.com/nyushi/barrow"
)

var (
	debug   = kingpin.Flag("debug", "Debug").Short('d').Bool()
	dryrun  = kingpin.Flag("dryrun", "Dryrun").Short('n').Bool()
	ruleDir = kingpin.Flag("ruledir", "Path to rule directory").Short('r').Default(".").String()
)

func main() {
	kingpin.Parse()
	rd, err := filepath.Abs(*ruleDir)
	if err != nil {
		logrus.Fatalf("invalid rule dir: %s", err)
	}
	rule, err := barrow.LoadRule(rd)
	if err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("dryrun mode enabled")
	if err := rule.Install(*dryrun); err != nil {
		logrus.Fatal(err)
	}
}
