package mongolog

import (
	"strings"
	"testing"
)

func TestParseCommandMessageValues(t *testing.T) {
	logLine := `2018-10-05T14:01:04.067+0000 I COMMAND  [conn206777]` +
		` command FooDb.mycatpicscollection command: find { find: "mycatpicscollection",` +
		` filter: { foo.FooObjectId: ObjectId('5a8c3a142053a407a936745e'), foo.max_time:` +
		` { $gte: 1534769530.5 }, foo.min_time: { $lte: 1534769548.47 }, foo.category:` +
		` { $in: [ "alley", "home" ] } }, $db: "FooDb" } planSummary: IXSCAN` +
		` { foo.FooObjectId: 1, foo.category: 1, foo.min_time: -1, foo.max_time: 1 }` +
		` keysExamined:50314 docsExamined:2 cursorExhausted:1 numYields:393 nreturned:2 reslen:14980` +
		` locks:{ Global: { acquireCount: { r: 788 } }, Database: { acquireCount: { r: 394 } },` +
		` Collection: { acquireCount: { r: 394 } } } protocol:op_query 219ms`

	matches := RegexpMatch(MongoLoglineRegex, logLine)
	expectValues := map[string]string{
		"timestamp": "2018-10-05T14:01:04.067+0000",
		"severity":  "I",
		"component": "COMMAND",
		"context":   "[conn206777]",
		"message":   "command FooDb",
	}

	for k, v := range expectValues {
		if k == "message" {
			if !strings.HasPrefix(matches[k], v) {
				t.Errorf("message data mismatch")
			}
		} else if v != matches[k] {
			t.Errorf("Expected %v='%v', got '%v'\n", k, v, matches[k])
		}
	}

	messageText := matches["message"]
	matches = RegexpMatch(MongoLogPayloadRegex, messageText)
	expectValues = map[string]string{
		"collection":    "FooDb.mycatpicscollection",
		"command":       "find",
		"commandparams": "{ find: \"mycatpicscollection\"",
		"plansummary":   "IXSCAN { foo.FooObjectId",
		"protocol":      "op_query",
		"duration":      "219",
	}

	for k, v := range expectValues {
		if !strings.HasPrefix(matches[k], v) {
			t.Errorf("Expected %v='%v', got '%v'\n", k, v, matches[k])
		}
	}
}
