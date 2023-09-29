package main

import (
	"fmt"
	"os"
)

func exitIfPanic(recovered any) {
	if recovered == nil {
		return
	}
	err := errorf("panic: %v", recovered)
	msg, _ := parseError(err)
	fmt.Println(":exclamationmark.triangle.fill:")
	fmt.Println("---")
	fmt.Println("Error:", msg)

	logError(err)
	_ = errWriter.Flush()
	os.Exit(1)
}

func exitWithError(err error) {
	msg, _ := parseError(err)
	fmt.Println(":exclamationmark.triangle.fill:")
	fmt.Println("---")
	fmt.Println("Error:", msg)

	logError(err)
	_ = errWriter.Flush()
	os.Exit(1)
}
