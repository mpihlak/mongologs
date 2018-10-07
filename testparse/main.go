package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mpihlak/mongolog"
)

func dumpContext(s string, m map[string]string) {
	fmt.Printf("context: %v\n", s)
	for k, v := range m {
		fmt.Printf("%v: %v\n", k, v)
	}
}

func main() {
	parser, err := mongolog.NewPseudoJsonParser()
	if err != nil {
		panic(err)
	}

	var file io.Reader

	if len(os.Args) == 2 {
		file, err = os.Open(os.Args[1])
		if err != nil {
			panic(err)
		}
	} else {
		file = os.Stdin
	}

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		logLine := scanner.Text()

		logMatch := mongolog.RegexpMatch(mongolog.MongoLoglineRegex, logLine)
		message := logMatch["message"]
		component := logMatch["component"]
		severity := logMatch["severity"]

		if severity != "I" || component != "COMMAND" {
			// Only look at "normal" command log
			continue
		}

		contentMatch := mongolog.RegexpMatch(mongolog.MongoLogPayloadRegex, message)

		if commandParams, ok := contentMatch["commandparams"]; ok {
			_, err = mongolog.ParseMessage(parser, commandParams)
			if err != nil {
				fmt.Printf("commandparams parse error: %v\n", err)
				dumpContext(commandParams, contentMatch)
			}
		} else {
			fmt.Printf("command parameters not found.\n")
			dumpContext(message, contentMatch)
		}

		if planSummary, ok := contentMatch["plansummary"]; ok {
			_, err = mongolog.ParseMessage(parser, planSummary)
			if err != nil {
				fmt.Printf("plansummary parse error: %v\n", err)
				dumpContext(planSummary, contentMatch)
			}
		} else {
			fmt.Printf("plansummary not found.\n")
			dumpContext(message, contentMatch)
		}

		fmt.Println()
	}
}
