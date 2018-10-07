package main

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mpihlak/mongolog"
)

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

		fmt.Printf("logLine: %v\n", logLine)

		match := mongolog.RegexpMatch(mongolog.MongoLoglineRegex, logLine)
		for k, v := range match {
			fmt.Printf("%v: %v\n", k, v)
		}

		message, ok := match["message"]
		if !ok {
			fmt.Printf("message body not found.\n\n")
			continue
		}

		match = mongolog.RegexpMatch(mongolog.MongoLogPayloadRegex, message)
		for k, v := range match {
			fmt.Printf("%v: %v\n", k, v)
		}

		if commandParams, ok := match["commandparams"]; ok {
			_, err = mongolog.ParseMessage(parser, commandParams)
			if err != nil {
				fmt.Printf("commandparams parse error: %v\n", err)
			}
		} else {
			fmt.Printf("command parameters not found.\n")
		}

		if planSummary, ok := match["plansummary"]; ok {
			_, err = mongolog.ParseMessage(parser, planSummary)
			if err != nil {
				fmt.Printf("plansummary parse error: %v\n", err)
			}
		} else {
			fmt.Printf("plansummary not found.\n")
		}

		fmt.Println()
	}
}
