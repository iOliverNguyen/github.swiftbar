package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

var verbosed = false
var errWriter *bufio.Writer
var logFile *os.File

func setupLogFile(path string) {
	must(0, os.MkdirAll(pathLogDir, 0755))
	if path == "" {
		errWriter = bufio.NewWriter(os.Stderr)
		return
	}
	if _, err := os.Stat(path); err == nil {
		must(0, os.Remove(path))
	}
	logFile = must(os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644))
	errWriter = bufio.NewWriter(logFile)
}

var regexpNumber = regexp.MustCompile(`[0-9]+`)

func parseNumber(s string) (int, bool) {
	parts := regexpNumber.FindStringSubmatch(s)
	if parts == nil {
		return 0, false
	}
	return must(strconv.Atoi(parts[0])), true
}

func get[T any](A []T, i int) (out T) {
	if len(A) <= i {
		return out
	}
	return A[i]
}

func xif[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func toGhURL(path string) string {
	if strings.HasPrefix(path, "https://") {
		return path
	}
	return "https://github.com" + path
}

func fprint(w io.Writer, args ...any) {
	_, err := fmt.Fprint(w, args...)
	if err != nil {
		panic(err)
	}
}

func fprintln(w io.Writer, args ...any) {
	_, err := fmt.Fprintln(w, args...)
	if err != nil {
		panic(err)
	}
}

func fprintf(w io.Writer, format string, args ...any) {
	_, err := fmt.Fprintf(w, format, args...)
	if err != nil {
		panic(err)
	}
}

func errorf(msg string, args ...any) error {
	msg = fmt.Sprintf(msg, args...)
	msg = msg + "\n\n" + string(debug.Stack())
	return errors.New(msg)
}

func wrapf(err error, msg string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%v: %v", fmt.Sprintf(msg, args...), err)
}

func parseError(err error) (msg, stack string) {
	if err == nil {
		return "", ""
	}
	parts := strings.Split(err.Error(), "\n\n")
	msg, stack = parts[0], ""
	if len(parts) > 0 {
		stack = parts[1]
	}
	return msg, stack
}

func logError(err error) {
	if err == nil || !verbosed {
		return
	}
	msg, stack := parseError(err)
	fprintf(errWriter, "[ERROR] "+msg+"\n")
	if stack != "" {
		fprintf(errWriter, "\n"+stack+"\n")
	}
}

func debugf(msg string, args ...any) {
	if verbosed {
		fprintf(errWriter, "[DEBUG] "+msg+"\n", args...)
	}
}

func debugYaml(v interface{}) {
	if verbosed {
		fprintf(errWriter, "%v\n", toDebugYaml(v))
	}
}

func exitf(msg string, args ...any) {
	if verbosed {
		fprintf(errWriter, "[FATAL] "+msg+"\n", args...)
	}
	os.Exit(1)
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func revert[T any](list []T) []T {
	out := make([]T, len(list))
	for i, v := range list {
		out[len(list)-i-1] = v
	}
	return out
}

func formatKey(key string) string {
	var b strings.Builder
	key = strings.ToLower(key)
	for i, word := range strings.Split(key, "-") {
		if i > 0 {
			b.WriteString("-")
		}
		if word == "" {
			continue
		}
		b.WriteString(strings.ToUpper(word[0:1]))
		b.WriteString(word[1:])
	}
	return b.String()
}

func execGit(args ...string) (string, error) {
	return execCommand("git", args...)
}

func execCommand(name string, args ...string) (string, error) {
	if verbosed {
		b := &strings.Builder{}
		fprint(b, name, " ")
		for _, arg := range args {
			if strings.Contains(arg, " ") {
				fprintf(b, "%q", arg)
			} else {
				fprint(b, arg, " ")
			}
		}
		debugf(b.String())
	}
	stdout := bytes.Buffer{}
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = &stdout, &stdout
	err := cmd.Run()
	if err != nil {
		err = errorf("%v %v: %v: %v", name, strings.Join(args, " "), err, stdout.String())
	}
	return stdout.String(), err
}

func getLineByPrefix(lines []string, prefix string) string {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func readFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func toJson(data interface{}) string {
	out, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func toYaml(v any) string {
	var b bytes.Buffer
	enc := yaml.NewEncoder(&b)
	enc.SetIndent(4)
	must(0, enc.Encode(v))
	return b.String()
}

func toDebugYaml(v any) string {
	return "\n" + strings.ReplaceAll(toYaml(v), "    ", ".   ")
}

func mergeLists[T any](lists ...[]T) []T {
	var out []T
	for _, list := range lists {
		out = append(out, list...)
	}
	return out
}

func appendUnique[T comparable](list []T, item T) []T {
	for _, v := range list {
		if v == item {
			return list
		}
	}
	return append(list, item)
}
