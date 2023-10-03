package main

import (
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

var topMergedPRsFreshTime = time.Now().Add(-1 * 24 * time.Hour)
var topOpenPRsFreshTime = time.Now().Add(-3 * 24 * time.Hour)
var myOtherPRsFreshTime = time.Now().Add(-7 * 24 * time.Hour)
var otherPRsMergedLatestTime = time.Now().Add(-1 * 24 * time.Hour)
var otherPRsOpenLatestTime = time.Now().Add(-3 * 24 * time.Hour)
var otherPRsFreshTime = time.Now().Add(-7 * 24 * time.Hour)

var checkLevels map[string]int
var importantChecks []string
