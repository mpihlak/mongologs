package mongolog

import (
	"testing"
)

func TestParseCommandMessage(t *testing.T) {
	logLine := ""
	logLine = `2018-10-05T14:01:04.067+0000 I COMMAND  [conn206777]` +
		` command FooDb.mycatpicscollection command: find { find: "mycatpicscollection",` +
		` filter: { foo.FooObjectId: ObjectId('5a8c3a142053a407a936745e'), foo.max_time:` +
		` { $gte: 1534769530.5 }, foo.min_time: { $lte: 1534769548.47 }, foo.category:` +
		` { $in: [ "alley", "home" ] } }, $db: "FooDb" } planSummary: IXSCAN` +
		` { foo.FooObjectId: 1, foo.category: 1, foo.min_time: -1, foo.max_time: 1 }` +
		` keysExamined:50314 docsExamined:2 cursorExhausted:1 numYields:393 nreturned:2 reslen:14980` +
		` locks:{ Global: { acquireCount: { r: 788 } }, Database: { acquireCount: { r: 394 } },` +
		` Collection: { acquireCount: { r: 394 } } } protocol:op_query 219ms`

	matches := RegexpMatch(MongoLoglineRegex, logLine)
	expectKeys := []string{"timestamp", "severity", "component", "context", "message"}

	for _, k := range expectKeys {
		if _, ok := matches[k]; !ok {
			t.Errorf("Logline: expected key not found: %v\n", k)
		}
	}

	messageText := matches["message"]
	matches = RegexpMatch(MongoLogPayloadRegex, messageText)
	expectKeys = []string{"collection", "command", "commandparams", "plansummary", "protocol", "duration"}

	for _, k := range expectKeys {
		if _, ok := matches[k]; !ok {
			t.Errorf("Message: expected key not found: %v\n", k)
		}
	}
}
