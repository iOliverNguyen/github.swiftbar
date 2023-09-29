package main

import (
	"golang.org/x/exp/slices"
	"regexp"
	"strconv"
	"time"
)

type Input struct {
	PRs []int
}

type Stack struct {
	TopPRs   []*PR
	MyOthers []*PR
	Others   []*PR
}

type PR struct {
	Number int
	Title  string
	State  string // open | closed | merged
	URL    string
	User   string
	Merged bool
	Unread bool

	UpdatedAt   time.Time
	HeadCommit  string
	MergeCommit string

	Checks  []*PRCheck
	Details *HtmlPRDetails
}

type HtmlPR struct {
	Number    int
	Unread    bool
	NComments int
}

type HtmlPRDetails struct {
	Number int

	NStackPRs int
	NCommits  int
	NComments int

	HasMyComments         bool // comments, approve or changes requested
	ApprovedUsers         []string
	ChangesRequestedUsers []string
	CommentsOnlyUsers     []string
}

type PRCheck struct {
	Name    string
	Status  string // pass | pending | fail | skipping
	StScore int    // pass | pending | fail (1 0 -1)
	URL     string

	Level int // 0 | 1 | 2
}

func (s *Stack) Get(number int) *PR {
	for _, pr := range s.TopPRs {
		if pr.Number == number {
			return pr
		}
	}
	return nil
}

func (pr *PR) GetCheck(name string) *PRCheck {
	for _, c := range pr.Checks {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func (pr *PR) MostChecks() (out []*PRCheck) {
	for _, c := range pr.Checks {
		if c.Level >= 1 {
			out = append(out, c)
		}
	}
	slices.SortFunc(out, func(a, b *PRCheck) int {
		return b.StScore - a.StScore
	})
	return out
}

func (pr *PR) ImportantChecks() (out []*PRCheck) {
	for _, name := range importantChecks {
		ch := pr.GetCheck(name)
		if ch == nil {
			ch = &PRCheck{
				Name: name,
			}
		}
		out = append(out, ch)
	}
	return out
}

func calcStatusScore(status string) int {
	switch status {
	case "pass":
		return 2
	case "pending":
		return 1
	case "skipping":
		return 0
	case "fail":
		return -1
	default:
		return 0
	}
}

func calcChecksLevel(checks []*PRCheck) {
	for _, c := range checks {
		c.Level = checkLevels[c.Name]
	}
}

func parseStack(s string) (out []int) {
	reNumber := regexp.MustCompile(`[0-9]+`)
	numbers := reNumber.FindAllString(s, -1)
	for _, number := range numbers {
		out = append(out, must(strconv.Atoi(number)))
	}
	return out
}
