package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) <= 1 {
		return
	}
	input := os.Args[1]
	lines := strings.Split(input, ": ")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "go")
		line = strings.TrimSpace(line)
		if line != "" {
			fmt.Println(strings.TrimSpace(line))
		}
	}
	t := time.Now()
	fmt.Printf("Last refresh at %02v:%02v | refresh=true\n", t.Hour(), t.Minute())
}
