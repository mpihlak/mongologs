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
	matches = RegexpMatch(MongoLogCommandPayloadRegex, messageText)
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

func checkExpectedValues(t *testing.T, expectedValues, actualValues map[string]string) {
	for k, v := range expectedValues {
		if actualValues[k] != v {
			t.Errorf("Expected %v='%v', got '%v'\n", k, v, actualValues[k])
		}
	}
}

func TestParseNewConnection(t *testing.T) {
	message := `connection accepted from 10.178.5.250:47878 #2078609 (252 connections now open)`
	matches := RegexpMatch(MongoNewConnectionRegex, message)

	expectValues := map[string]string{
		"ip":   "10.178.5.250",
		"port": "47878",
		"id":   "2078609",
	}
	checkExpectedValues(t, expectValues, matches)
}

func TestParseEndConnection(t *testing.T) {
	message := `end connection 127.0.0.1:42266 (250 connections now open)`
	matches := RegexpMatch(MongoEndConnectionRegex, message)

	expectValues := map[string]string{
		"ip":   "127.0.0.1",
		"port": "42266",
	}
	checkExpectedValues(t, expectValues, matches)
}

func TestParseConnectionMetadata(t *testing.T) {
	message := `received client metadata from 10.178.5.250:47876 conn2078608: { driver: { name: "PyMongo" } }`
	matches := RegexpMatch(MongoConnectionMetadataRegex, message)

	expectValues := map[string]string{
		"ip":       "10.178.5.250",
		"port":     "47876",
		"id":       "conn2078608",
		"metadata": `{ driver: { name: "PyMongo" } }`,
	}
	checkExpectedValues(t, expectValues, matches)
}

func TestParseInsertCommand(t *testing.T) {
	message := `command FooDb.mycatpicscollection command: insert` +
		` { insert: "mycatpicscollection", ordered: true, $clusterTime:` +
		` { clusterTime: Timestamp(1538979514, 76), signature: {` +
		` hash: BinData(0, 0000000000000000000000000000000000000000), keyId: 0 } },` +
		` lsid: { id: UUID("c3cc9fef-182a-4917-9b5a-f715d0639ac2") }, $db: "FooDb" }` +
		` ninserted:1 keysInserted:3 numYields:0 reslen:229 locks:{ Global: ` +
		`{ acquireCount: { r: 2, w: 2 } }, Database: { acquireCount: { w: 2 },` +
		` acquireWaitCount: { w: 1 }, timeAcquiringMicros: { w: 12259 } },` +
		` Collection: { acquireCount: { w: 1 } }, oplog: { acquireCount: { w: 1 } } }` +
		` protocol:op_query 12ms`

	matches := RegexpMatch(MongoLogInsertPayloadRegex, message)

	expectValues := map[string]string{
		"command":    "insert",
		"collection": "FooDb.mycatpicscollection",
		"protocol":   "op_query",
		"duration":   "12",
	}
	checkExpectedValues(t, expectValues, matches)
}

func TestReplaceBinData(t *testing.T) {
	message := `{ a: BinData(0, E3B0C44298FC1), b: BinData(0, E3B0C44298FC1) }`
	expectMessage := `{ a: BinData(0, "E3B0C44298FC1"), b: BinData(0, "E3B0C44298FC1") }`

	s := ReplaceBinData(message)
	if s != expectMessage {
		t.Errorf("unexpected result:\nexpect: %v\nvalue : %v\n", expectMessage, s)
	}
}
