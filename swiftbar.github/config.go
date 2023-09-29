package main

import (
	"fmt"
	"time"
)

var githubAPIPath = "https://api.github.com/repos/{ORG}/{REPO}"
var githubHtmlPath = "https://github.com/{ORG}/{REPO}"

var pathLogDir = ""
var pathPrsTxt = ""
var pathErrLog = ""

var githubUser = ""
var githubToken = ""
var githubCookie = ""
var htmlHeaders []string

var timeout = 10 * time.Minute

var topOpenPRsFreshTime = time.Now().Add(-3 * 24 * time.Hour)
var topMergedPRsFreshTime = time.Now().Add(-1 * 24 * time.Hour)
var myOtherPRsFreshTime = time.Now().Add(-7 * 24 * time.Hour)
var otherPRsLatestTime = time.Now().Add(-3 * 24 * time.Hour)
var otherPRsFreshTime = time.Now().Add(-7 * 24 * time.Hour)

var checkLevels = map[string]int{
	"build (>=1.19.2, go)":                   2,
	"code-gen":                               1,
	"integration-test (go, >=1.19.2)":        1,
	"proto-generation":                       1,
	"race-test (go, >=1.19.2)":               1,
	"unit-test (go, >=1.19.2)":               2,
	"Linter":                                 0,
	"build (14.x)":                           0, // chat-server
	"build (>=1.19.2, go/tenets)":            2,
	"changes":                                0,
	"idl-vendor-check":                       0,
	"integration-test (go/tenets, >=1.19.2)": 1,
	"lint (>=1.18.0, ubuntu-20.04)":          2,
	"proto-breaking":                         1,
	"proto-lint":                             2,
	"race-test (go/tenets, >=1.19.2)":        1,
	"unit-test (go/tenets, >=1.19.2)":        2,
}

var importantChecks = []string{
	"lint (>=1.18.0, ubuntu-20.04)",
	"build (>=1.19.2, go)",
	"unit-test (go, >=1.19.2)",
	"build (>=1.19.2, go/tenets)",
	"unit-test (go/tenets, >=1.19.2)",
	"proto-lint",
}

func init() {
	for _, k := range importantChecks {
		level, ok := checkLevels[k]
		if !ok {
			panic(fmt.Sprintf("not found %q", k))
		}
		if level != 2 {
			panic(fmt.Sprintf("%q: level=%v", k, level))
		}
	}
}
