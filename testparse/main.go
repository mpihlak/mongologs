package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mpihlak/mongolog"
)

func dumpContext(s string, m map[string]string) {
	fmt.Printf("context: %v\n", s)
	for k, v := range m {
		fmt.Printf("%v: %v\n", k, v)
	}
	fmt.Println()
}

func main() {
	file := os.Stdin

	if len(os.Args) == 2 {
		var err error
		file, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
		defer file.Close()
	}

	scanner := bufio.NewScanner(file)
	parser, err := mongolog.NewLogParser()
	if err != nil {
		panic(err)
	}

	total_lines := 0
	parse_errors := 0
	for scanner.Scan() {
		logLine := scanner.Text()
		total_lines++

		logEntry, err := mongolog.ParseLogEntry(parser, logLine)
		if err != nil {
			fmt.Printf("error parsing: %v\n", logLine)
			fmt.Printf("%s\n", err)
			parse_errors++
		} else {
			chop := len(logEntry.LogMessage)
			if chop > 80 {
				chop = 80
			}

			/*
				fmt.Printf("time: %v\nseverity: %v\ncomponent: %v\ncontext: %v\nlog: %v\n\n",
					logEntry.Timestamp, logEntry.Severity, logEntry.Component,
					logEntry.Context, logEntry.LogMessage[:chop])
			*/
		}
	}

	fmt.Printf("Done, total lines %d, parse errors %d\n", total_lines, parse_errors)
}
