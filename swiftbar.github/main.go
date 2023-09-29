package main

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	"sync"
)

func main() {
	flag.BoolVar(&verbosed, "v", false, "verbose")
	flag.Parse()
	setupLogFile(xif(verbosed, pathErrLog, ""))

	var prNumbers []int
	dataPrsTxt, err := readFile(pathPrsTxt)
	debugf("read prs file: %v", xif(err == nil, "ok", fmt.Sprint(err)))
	if err == nil {
		prNumbers = parseStack(dataPrsTxt)
	}
	input := &Input{PRs: prNumbers}
	stack := mainExec(input)
	debugf("--- STACK ---")
	debugYaml(stack)

	renderPR := func(pr *PR) {
		statusImg := imageToBase64(renderPRImage(pr))
		extraStr := calcExtraStr(pr)
		fmt.Printf("%v (#%v) %v | href=%v image=%s\n", pr.Title, pr.Number, extraStr, pr.URL, statusImg)
		for _, check := range pr.Checks {
			status :=
				xif(check.Status == "pending", "🟡",
					xif(check.Status == "pass", "✅",
						xif(check.Status == "fail", "❌", "◽️")))
			fmt.Printf("--%v %v | href=%v\n", status, check.Name, check.URL)
		}
	}

	statusImg := imageToBase64(renderStackImage(stack))
	fmt.Printf("| image=%s\n", statusImg)
	fmt.Println("---")
	for _, pr := range stack.TopPRs {
		renderPR(pr)
	}
	if len(stack.MyOthers) > 0 {
		fmt.Println("---")
	}
	for _, pr := range stack.MyOthers {
		renderPR(pr)
	}
	if len(stack.Others) > 0 {
		fmt.Println("---")
	}
	for _, pr := range stack.Others {
		renderPR(pr)
	}

	if errWriter != nil {
		must(0, errWriter.Flush())
	}
}

func mainExec(input *Input) *Stack {
	recentPRs := listGhPRs()
	myTopPRs := filterStackedPRs(recentPRs, input.PRs)
	if len(myTopPRs) < 3 { // less than 3 PRs, include my recent top PRs
		myRecentPRs := sortPRs(skipPRs(filterMyRecentTopPRs(recentPRs), myTopPRs))
		myTopPRs = mergeLists(myTopPRs, myRecentPRs)
	}
	if len(myTopPRs) < 3 { // still less than 3 PRs, include my other PRs, max 3 PRs
		myRecentPRs := sortPRs(skipPRs(filterMyOtherPRs(recentPRs), myTopPRs))
		mergeLists(myTopPRs, myRecentPRs)
		if len(myTopPRs) > 3 {
			myTopPRs = myTopPRs[:3]
		}
	}
	myOthers := skipPRs(filterMyOtherPRs(recentPRs), myTopPRs)
	stack := &Stack{TopPRs: myTopPRs, MyOthers: myOthers}
	{
		wg := sync.WaitGroup{}
		wg.Add(4)
		result := make([][]*HtmlPR, 4)
		for page := 0; page < 4; page++ {
			page := page
			go func() {
				defer func() { exitIfPanic(recover()) }()
				result[page] = listGhPRsHtml(page)
				wg.Done()
			}()
		}
		wg.Wait()
		stack.Others = filterOtherPRs(recentPRs, result, stack.TopPRs, stack.MyOthers)
	}
	{
		wg := sync.WaitGroup{}
		allPRs := mergeLists(stack.TopPRs, stack.MyOthers, stack.Others)
		for _, pr := range allPRs {
			pr := pr
			wg.Add(1)
			go func() {
				defer func() { exitIfPanic(recover()) }()
				if pr == nil {
					pr = queryGhPR(pr.Number)
				}
				pr.Checks = queryGhPRChecksHtml(pr.Number)
				pr.Details = queryGhPRHtml(pr.Number)
				wg.Done()
			}()
		}
		wg.Wait()
	}
	return stack
}

func sortPRs(prs []*PR) []*PR {
	sort.Slice(prs, func(i, j int) bool {
		return prs[i].UpdatedAt.After(prs[j].UpdatedAt)
	})
	return prs
}

func skipPRs(prs []*PR, skips []*PR) (out []*PR) {
	mapSkipPRs := map[int]bool{}
	for _, pr := range skips {
		mapSkipPRs[pr.Number] = true
	}
	for _, pr := range prs {
		if !mapSkipPRs[pr.Number] {
			out = append(out, pr)
		}
	}
	return out
}

func filterStackedPRs(recentPRs []*PR, stackedPRs []int) (out []*PR) {
	mapPR := map[int]*PR{}
	for _, pr := range recentPRs {
		mapPR[pr.Number] = pr
	}
	for _, number := range stackedPRs {
		pr := mapPR[number]
		if pr != nil {
			out = append(out, mapPR[number])
		}
	}
	return out
}

func filterMyRecentTopPRs(recentPRs []*PR) (out []*PR) {
	for _, pr := range recentPRs {
		switch {
		case pr.User == githubUser && pr.State == "open" && pr.UpdatedAt.After(topOpenPRsFreshTime):
			out = append(out, pr)
		case pr.User == githubUser && pr.Merged && pr.UpdatedAt.After(topMergedPRsFreshTime):
			out = append(out, pr)
		}
	}
	return out
}

func filterMyOtherPRs(recentPRs []*PR) (out []*PR) {
	for _, pr := range recentPRs {
		ok := pr.State == "open" || pr.Merged
		if ok && pr.User == githubUser && pr.UpdatedAt.After(myOtherPRsFreshTime) {
			out = append(out, pr)
		}
	}
	return out
}

func filterOtherPRs(recentPRs []*PR, htmlPRs [][]*HtmlPR, excludes ...[]*PR) (out []*PR) {
	mapExcludedPRs := map[int]bool{}
	for _, exclude := range excludes {
		for _, pr := range exclude {
			mapExcludedPRs[pr.Number] = true
		}
	}
	mapHtmlPRs := map[int]*HtmlPR{}
	for _, page := range htmlPRs {
		for _, pr := range page {
			mapHtmlPRs[pr.Number] = pr
		}
	}
	debugf("--- HTML PRs ---")
	debugYaml(mapHtmlPRs)
	for _, pr := range recentPRs {
		if mapExcludedPRs[pr.Number] {
			continue
		}
		htmlPR := mapHtmlPRs[pr.Number]
		if htmlPR == nil {
			continue
		}
		pr.Unread = htmlPR.Unread
		switch {
		case pr.UpdatedAt.After(otherPRsLatestTime):
			out = append(out, pr)
		case htmlPR.NComments > 0 && pr.UpdatedAt.After(otherPRsFreshTime):
			out = append(out, pr)
		}
	}
	return out
}

func calcExtraStr(pr *PR) string {
	approval := 0
	if len(pr.Details.ApprovedUsers) > 0 {
		approval = 1
	}
	if len(pr.Details.ChangesRequestedUsers) > 0 {
		approval = -1
	}

	b := &strings.Builder{}
	switch {
	case approval == 1 && pr.Unread:
		fprintf(b, ":checkmark.seal.fill: ")
	case approval == 1 && !pr.Unread:
		fprintf(b, ":checkmark.seal.fill: ")
	case approval == -1 && pr.Unread:
		fprintf(b, ":minus.circle: ")
	case approval == -1 && !pr.Unread:
		fprintf(b, ":minus.circle: ")
	}
	if pr.Details.NComments > 0 {
		fprintf(b, "  :bubble.right.fill:%v ", pr.Details.NComments)
	}
	if pr.Unread {
		fprintf(b, "⏺ ")
	}
	return b.String()
}
