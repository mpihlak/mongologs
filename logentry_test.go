package mongolog

import (
	"testing"
)

func TestParseLogEntry(t *testing.T) {
	newConnectionMessage := `2018-10-05T14:01:04.067+0000 I NETWORK  [listener]` +
		` connection accepted from 10.178.5.250:47878 #2078609 (252 connections now open)`
	clientMetadataMessage := `2018-10-05T14:01:04.067+0000 I NETWORK  [conn2078609]` +
		` received client metadata from 10.178.5.250:47878 conn2078609: { driver: { name: "PyMongo" } }`
	commandMessage := `2018-10-05T14:01:04.067+0000 I COMMAND  [conn2078609]` +
		` command FooDb.mycatpicscollection command: find { find: "mycatpicscollection",` +
		` filter: { foo.FooObjectId: ObjectId('5a8c3a142053a407a936745e'), foo.max_time:` +
		` { $gte: 1534769530.5 }, foo.min_time: { $lte: 1534769548.47 }, foo.category:` +
		` { $in: [ "alley", "home" ] } }, $db: "FooDb" } planSummary: IXSCAN` +
		` { foo.FooObjectId: 1, foo.category: 1, foo.min_time: -1, foo.max_time: 1 }` +
		` keysExamined:50314 docsExamined:2 cursorExhausted:1 numYields:393 nreturned:2 reslen:14980` +
		` locks:{ Global: { acquireCount: { r: 788 } }, Database: { acquireCount: { r: 394 } },` +
		` Collection: { acquireCount: { r: 394 } } } protocol:op_query 219ms`
	endConnectionMessage := `2018-10-05T14:01:04.067+0000 I NETWORK  [conn2078609]` +
		` end connection 127.0.0.1:42266 (250 connections now open)`

	validateConnection := func(m MongoLogEntry, err error) {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		if m.ConnectionInfo == nil {
			t.Errorf("expected connection info not there")
			return
		}
		if m.ConnectionInfo.ConnectionId != "[conn2078609]" {
			t.Errorf("unexpected connection id: %v", m.ConnectionInfo.ConnectionId)
		}
	}

	parser, _ := NewLogParser()

	m, err := ParseLogEntry(parser, newConnectionMessage)
	validateConnection(m, err)

	m, err = ParseLogEntry(parser, clientMetadataMessage)
	validateConnection(m, err)

	m, err = ParseLogEntry(parser, commandMessage)
	validateConnection(m, err)

	m, err = ParseLogEntry(parser, endConnectionMessage)
	validateConnection(m, err)
}

