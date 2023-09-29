package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"regexp"
	"strings"
	"time"
)

type PRResponse struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	State       string `json:"state"`
	URL         string `json:"html_url"`
	MergeCommit string `json:"merge_commit_sha"`
	Head        struct {
		Sha string `json:"sha"`
	} `json:"head"`
	User struct {
		Login string `json:"login"`
	}
	UpdatedAt time.Time `json:"updated_at"`

	MergedAt time.Time `json:"merged_at"` // only merged or merged_at
	Merged   bool      `json:"merged"`    //
}

func (x *PRResponse) GetMerged() bool {
	return x.Merged || !x.MergedAt.IsZero()
}

func parsePRResponse(data PRResponse) *PR {
	return &PR{
		Number:      data.Number,
		Title:       data.Title,
		State:       data.State,
		URL:         data.URL,
		Merged:      data.GetMerged(),
		HeadCommit:  data.Head.Sha,
		MergeCommit: data.MergeCommit,
		User:        data.User.Login,
		UpdatedAt:   data.UpdatedAt,
	}
}

func queryGhPR(number int) *PR {
	ghURL := fmt.Sprintf("%v/pulls/%v", githubAPIPath, number)
	jsonBody := must(httpGET(ghURL))

	var res PRResponse
	must(0, json.Unmarshal(jsonBody, &res))
	return parsePRResponse(res)
}

func listGhPRs() (out []*PR) {
	ghURL := fmt.Sprintf("%v/pulls?state=all&per_page=100", githubAPIPath)
	jsonBody := must(httpRequest("GET", ghURL, nil))

	var response []PRResponse
	must(0, json.Unmarshal(jsonBody, &response))
	for _, res := range response {
		pr := parsePRResponse(res)
		out = append(out, pr)
	}
	debugf("--- PRs ---")
	debugYaml(out)
	return out
}

var regexpUser = regexp.MustCompile(`^[A-Za-z0-9-]+$`)

func queryGhPRHtml(number int) *HtmlPRDetails {
	ghURL := fmt.Sprintf("%v/pull/%v", githubHtmlPath, number)
	htmlBody := must(htmlRequest(ghURL))
	doc := must(goquery.NewDocumentFromReader(bytes.NewReader(htmlBody)))
	if number == 4793 {
		debugf(string(htmlBody))
	}

	out := &HtmlPRDetails{Number: number}
	out.NStackPRs = doc.Find(`.comment-body [href*='/pull/']`).Length()
	out.NCommits, _ = parseNumber(doc.Find(`#commits_tab_counter`).First().Text())
	out.NComments, _ = parseNumber(doc.Find(`#conversation_tab_counter`).First().Text())
	// parse approved & changes requested users, parse HasMyComments
	doc.Find(`.sidebar-assignee .d-flex`).Each(func(i int, selection *goquery.Selection) {
		user := strings.TrimSpace(selection.Find(`a.assignee`).First().Text())
		if !regexpUser.MatchString(user) {
			return
		}
		hasComment := false
		selection.Find(`svg`).Each(func(i int, selection *goquery.Selection) {
			node := WrapNode(selection.Nodes[0])
			class := node.GetAttr("class")
			if strings.Contains(class, "success") {
				hasComment = true
				out.ApprovedUsers = appendUnique(out.ApprovedUsers, user)
			}
			if strings.Contains(class, "danger") {
				hasComment = true
				out.ChangesRequestedUsers = appendUnique(out.ChangesRequestedUsers, user)
			}
			if strings.Contains(class, "muted") {
				hasComment = true
				out.CommentsOnlyUsers = appendUnique(out.CommentsOnlyUsers, user)
			}
		})
		if hasComment && user == githubUser {
			out.HasMyComments = true
		}
	})
	return out
}

func listGhPRsHtml(page int) (out []*HtmlPR) {
	ghURL := fmt.Sprintf("%v/pulls", githubHtmlPath)
	if page > 0 {
		ghURL += fmt.Sprintf("?page=%v", page+1)
	}
	htmlBody := must(htmlRequest(ghURL))
	doc := must(goquery.NewDocumentFromReader(bytes.NewReader(htmlBody)))

	doc.Find(`.js-issue-row`).Each(func(i int, selection *goquery.Selection) {
		node := WrapNode(selection.First().Nodes[0])
		if node == nil {
			panic("failed to find node .js-issue-row")
		}
		number, ok := parseNumber(node.GetAttr("id")) // issue_1234
		if !ok {
			return
		}
		// parse unread
		unread := len(selection.Filter(`[class*=unread]`).Nodes) > 0
		// parse number of comments
		var commentsSelection *goquery.Selection
		selection.Find("a.Link--muted").Each(func(i int, selection *goquery.Selection) {
			if len(selection.Find("svg").Nodes) > 0 {
				commentsSelection = selection
			}
		})
		nComments := 0
		if commentsSelection != nil {
			text := commentsSelection.Find(".text-small").Text()
			nComments, _ = parseNumber(text)
		}
		pr := &HtmlPR{
			Number:    number,
			Unread:    unread,
			NComments: nComments,
		}
		out = append(out, pr)
	})
	return out
}

func queryGhPRChecksHtml(number int) (out []*PRCheck) {
	ghURL := fmt.Sprintf("%v/pull/%v/checks", githubHtmlPath, number)
	htmlBody := must(htmlRequest(ghURL))
	doc := must(goquery.NewDocumentFromReader(bytes.NewReader(htmlBody)))

	doc.Find(`.checks-list-item`).Each(func(i int, selection *goquery.Selection) {
		linkSel := selection.Find("a").First()
		if linkSel.Length() == 0 {
			return
		}
		node := WrapNode(linkSel.Nodes[0])
		href := node.GetAttr("href")
		if !regexpRun.MatchString(href) {
			return
		}
		name := strings.TrimSpace(linkSel.Text())

		status := ""
		switch {
		case selection.Find(`[class*=success]`).Length() > 0:
			status = "pass"
		case selection.Find(`[class*=danger]`).Length() > 0:
			status = "fail"
		case selection.Find(`svg[class*=rotate]`).Length() > 0:
			status = "pending"
		default:
			status = "skipping"
		}

		check := &PRCheck{
			Name:    name,
			Status:  status,
			StScore: calcStatusScore(status),
			URL:     toGhURL(href),
		}
		out = append(out, check)
	})
	calcChecksLevel(out)
	return out
}

var regexpRun = regexp.MustCompile(`actions/runs/(\d+)/job/`)

type Node html.Node

func WrapNode(n *html.Node) *Node {
	return (*Node)(n)
}

func (n Node) GetAttr(key string) string {
	for _, kv := range n.Attr {
		if kv.Key == key {
			return kv.Val
		}
	}
	return ""
}
