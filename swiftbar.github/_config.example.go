package main

// copy this file to config.local.go and change the values
func init() {
	githubAPIPath = "https://api.github.com/repos/{ORG}/{REPO}"
	githubHtmlPath = "https://github.com/{ORG}/{REPO}"

	pathLogDir = "PATH/TO/github.swiftbar/__/"
	pathPrsTxt = "PATH/TO/github.swiftbar/__/.prs.txt"
	pathErrLog = "PATH/TO/github.swiftbar/__/error.log"

	githubUser = "iOliverNguyen"
	githubToken = "gho_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	githubCookie = "_octo=xxxxx; user_session=xxxxx; _gh_sess=xxxxx"

	htmlHeaders = []string{
		`authority: github.com`,
		`accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7`,
		`accept-language: en-US,en;q=0.9,vi;q=0.8`,
		`cache-control: max-age=0`,
		`cookie:` + githubCookie,
		`if-none-match: W/"7a8e3d57f52d9bb40bc8b327f004bcd8"`,
		`sec-ch-ua: "Chromium";v="116", "Not)A;Brand";v="24", "Google Chrome";v="116"`,
		`sec-ch-ua-mobile: ?0`,
		`sec-ch-ua-platform: "macOS"`,
		`sec-fetch-dest: document`,
		`sec-fetch-mode: navigate`,
		`sec-fetch-site: same-origin`,
		`sec-fetch-user: ?1`,
		`upgrade-insecure-requests: 1`,
		`user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36`,
	}
	checkLevels = map[string]int{
		"changes":                                   0,
		"code-gen":                                  1,
		"build (1.21, go)":                          2,
		"build (1.21, go/tenets)":                   2,
		"unit-test (1.21, go)":                      2,
		"unit-test (1.21, go/tenets)":               2,
		"integration-test (1.21, go)":               1,
		"integration-test (1.21, go/tenets)":        1,
		"unit-test-race-detector (1.21, go)":        1,
		"unit-test-race-detector (1.21, go/tenets)": 1,
		"proto-lint":                                2,
		"proto-breaking":                            1,
		"proto-generation":                          1,
		"Test":                                      0, // python
		"Lint Code Base":                            0,
		"build (14.x)":                              0, // chat-server
		"lint (1.21, ubuntu-latest)":                2,
	}
	importantChecks = []string{
		"lint (1.21, ubuntu-latest)",
		"build (1.21, go)",
		"unit-test (1.21, go)",
		"build (1.21, go/tenets)",
		"unit-test (1.21, go/tenets)",
		"proto-lint",
	}
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
