package main

import (
	"regexp"

	"github.com/mpihlak/mongolog"
)

func main() {
	logLine := `2018-10-05T14:01:04.067+0000 I COMMAND  [conn206777] command FooDb.mycatpicscollection command: find { find: "mycatpicscollection", filter: { foo.FooObjectId: ObjectId('5a8c3a142053a407a936745e'), foo.max_time: { $gte: 1534769530.5 }, foo.min_time: { $lte: 1534769548.47 }, foo.category: { $in: [ "alley", "home" ] } }, $db: "FooDb" } planSummary: IXSCAN { foo.FooObjectId: 1, foo.category: 1, foo.min_time: -1, foo.max_time: 1 } keysExamined:50314 docsExamined:2 cursorExhausted:1 numYields:393 nreturned:2 reslen:14980 locks:{ Global: { acquireCount: { r: 788 } }, Database: { acquireCount: { r: 394 } }, Collection: { acquireCount: { r: 394 } } } protocol:op_query 219ms`

	parser, err := mongolog.NewPseudoJsonParser()
	if err != nil {
		panic(err)
	}

	loglineRe := regexp.MustCompile(
		`(?P<timestamp>[^\s]+)\s` +
			`(?P<severity>.)\s` +
			`(?P<component>[^\s]+)\s+` +
			`(?P<context>[^\s]+)\s` +
			`(?P<message>.*)`)
	payloadRe := regexp.MustCompile(
		`command (?P<collection>[^\s]+)\scommand:\s` +
			`(?P<command>[^\s]+)\s` +
			`(?P<commandparams>{.*})\splanSummary:\s` +
			`(?P<planmethod>[A-Z]+)\s` +
			`(?P<planinfo>{.*})\sprotocol:` +
			`(?P<protocol>[^\s]+)\s` +
			`(?P<duration>[0-9]+)ms`)

	for i := 0; i < 1000000; i++ {
		match := loglineRe.FindStringSubmatch(logLine)
		messageBody := match[5]

		match = payloadRe.FindStringSubmatch(messageBody)
		commandParams := match[3]
		planInfo := match[5]

		_, err = mongolog.ParseMessage(parser, planInfo)
		if err != nil {
			panic(err)
		}

		_, err = mongolog.ParseMessage(parser, commandParams)
		if err != nil {
			panic(err)
		}
	}
}
