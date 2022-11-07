package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/atotto/clipboard"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

var jiraTicketFlag string

var templatesPath string

const env = "GO_PR_PATH"

func main() {
	l := log.New(os.Stderr, "gopr logger:\t", log.Ldate|log.Ltime|log.Lshortfile)

	if templatesPath = os.Getenv(env); templatesPath == "" {
		l.Fatalf("environment variable %q for PR templates is not present in environment", env)
	}

	out, err := parseFlags("gopr", os.Args[1:])
	if err == flag.ErrHelp {
		l.Println(out)
		os.Exit(2)
	} else if err != nil {
		l.Println("output:\n", out)
		os.Exit(1)
	}

	names, err := getTemplateNames(templatesPath)
	if err != nil {
		l.Fatalln(err)
	}

	templatesIndexes := createConsoleOutputForTemplateNames(names)

	fmt.Println("Please select the template index you want to use: ")
	fmt.Println(templatesIndexes)
	fmt.Print("index number: ")

	scanner := bufio.NewScanner(os.Stdin)
	var index int
	var isIdxValid bool
	for !isIdxValid {
		scanner.Scan()
		idx, err := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if err != nil {
			fmt.Println("the index must be a number please try again")
			fmt.Print("index number: ")
			continue
		}
		if idx <= 0 {
			fmt.Println("index cannot be a number less or equal than zero")
			fmt.Print("index number: ")
			continue
		}
		if x := len(names); idx > x {
			fmt.Println("index out of range")
			fmt.Print("index number: ")
			continue
		}
		isIdxValid = true
		index = idx
	}

	tpl, err := template.ParseFiles(templatesPath + "/" + names[index-1])
	if err != nil {
		l.Fatalln(err)
	}

	buf := &bytes.Buffer{}

	err = tpl.ExecuteTemplate(buf, names[index-1], map[string]string{"jira": jiraTicketFlag})
	if err != nil {
		l.Fatalln(err)
	}

	err = clipboard.WriteAll(buf.String())
	if err != nil {
		l.Fatalln(err)
	}

	fmt.Println("Done!")
}

func parseFlags(programName string, args []string) (string, error) {
	var buf bytes.Buffer

	flags := flag.NewFlagSet(programName, flag.ContinueOnError)

	flags.SetOutput(&buf)

	flags.StringVar(&jiraTicketFlag, "jira", "", "jira ticket number")

	err := flags.Parse(args)
	if err != nil {
		return buf.String(), err
	}

	return buf.String(), nil
}

func getTemplateNames(p string) ([]string, error) {
	var names []string

	skipDir := filepath.Base(p)

	err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() && d.Name() != skipDir {
			return filepath.SkipDir
		}

		if !d.IsDir() && strings.Contains(path, ".gohtml") {
			names = append(names, d.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(names) == 0 {
		return nil, fmt.Errorf("there are templates in the given path: %s", p)
	}

	return names, nil
}

func createConsoleOutputForTemplateNames(names []string) string {
	s := strings.Builder{}
	for i, n := range names {
		s.WriteString(fmt.Sprintf("(%d) - %s\n", i+1, n))
	}
	return s.String()
}
