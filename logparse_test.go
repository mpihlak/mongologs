package mongolog

import (
	"regexp"
	"testing"
)

func TestParseCommandParameters(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{
		count.x: "mycatpicscollection",
		query: {
			MyObjectId: ObjectId('5a2fc7bd9b45c7117bee26c5'),
			baz.max_time: { $gte: 1523022862.698 },
			fooLimit: 42,
			category: "bagfoo"
		},
		$readPreference: {
			mode: "secondaryPreferred"
		},
		$db: "FooDb"
	}`

	msg, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("unable to parse message: %v: %v\n", testMessage, err)
	}

	var expectedStringValues = map[string]string{
		"count.x": "mycatpicscollection",
		"$db":     "FooDb",
	}

	for k, v := range expectedStringValues {
		if msg.elems[k].StringValue != v {
			t.Errorf("Expected: %v: to be '%v', got '%v'", k, v, msg.elems[k].StringValue)
		}
	}

	// Look at "query" fields
	q := msg.elems["query"].Nested
	botstring := q.elems["MyObjectId"].ObjectIdValue
	if botstring != "5a2fc7bd9b45c7117bee26c5" {
		t.Errorf("query MyObjectId mismatch, got %v", botstring)
	}
	if q.elems["category"].StringValue != "bagfoo" {
		t.Errorf("query category mismatch, got %v", q.elems["category"].StringValue)
	}
	bazMaxTime := q.elems["baz.max_time"].Nested.elems["$gte"].NumericValue
	if bazMaxTime != 1523022862.698 {
		t.Errorf("baz.max_time mismatch, got %v", bazMaxTime)
	}
	fooLimit := q.elems["fooLimit"].NumericValue
	if fooLimit != 42 {
		t.Errorf("fooLimit mismatch, got %v", fooLimit)
	}

	readPreference := msg.elems["$readPreference"].Nested.elems["mode"].StringValue
	if readPreference != "secondaryPreferred" {
		t.Errorf("$readPreference mismatch, got %v", readPreference)
	}
}

func TestParseIdentValues(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ find: "foocollection",` +
		` id: UUID("a2c81b1a-4d36-45db-93f0-c0beefb77838"),` +
		` objectId: ObjectId("foo!") }`

	_, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("parse error: %v: %v\n", testMessage, err)
	}
}

func TestParseBoolValues(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ find: "foocollection", projection: {}, foo: 2 }`
	_, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("parse error: %v: %v\n", testMessage, err)
	}
}

func TestParseEmptyStructs(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ find: true, projection: false }`
	_, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("parse error: %v: %v\n", testMessage, err)
	}
}

func TestParseAdHocComplexThing(t *testing.T) {
	parser, err := NewPlanSummaryParser()
	if err != nil {
		t.Errorf("error creating parser: %v\n", err)
	}

	testMessage := `IXSCAN { a: 1, b: 1 }, IXSCAN { a: 1, b: 1 }`
	_, err = ParsePlanSummary(parser, testMessage)
	if err != nil {
		t.Errorf("parse error: %v: %v\n", testMessage, err)
	}
}

func TestParseAdHocComplexThing2(t *testing.T) {
	parser, err := NewPlanSummaryParser()
	if err != nil {
		t.Errorf("error creating parser: %v\n", err)
	}

	testMessage := `COLLSCAN keysExamined:0 docsExamined:45227 cursorExhausted:1 numYields:353 nreturned:0 reslen:140 locks:{ Global: { acquireCount: { r: 354 } }, Database: { acquireCount: { r: 354 } }, Collection: { acquireCount: { r: 354 } } }`
	_, err = ParsePlanSummary(parser, testMessage)
	if err != nil {
		t.Errorf("parse error: %v: %v\n", testMessage, err)
	}
}

func TestParseCommandParametersNegativeNumbers(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ a: -1, b: 2 }`

	msg, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("unable to parse message: %v: %v\n", testMessage, err)
	}

	a := msg.elems["a"].NumericValue
	if a != -1 {
		t.Errorf("Expecting -1, got %v\n", a)
	}

	b := msg.elems["b"].NumericValue
	if b != 2 {
		t.Errorf("Expecting 2, got %v\n", b)
	}
}

func TestParseCommandParametersArrayValues(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ a: [ -42, 55, 9 ] }`

	msg, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("unable to parse message: %v: %v\n", testMessage, err)
	}

	expectValues := []float64{-42, 55, 9}
	for pos, v := range expectValues {
		a := msg.elems["a"].ArrayValue[pos].NumericValue
		if a != v {
			t.Errorf("Expecting %v, got %v\n", v, a)
		}
	}
}

func TestParseCommandParametersMixedMode(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `x: 1  y: 2 { z: 3 }`

	msg, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("unable to parse message: %v: %v\n", testMessage, err)
	}

	var expectedValues = map[string]float64{
		"x": 1, "y": 2, "z": 3,
	}
	for k, v := range expectedValues {
		actualValue := msg.elems[k].NumericValue
		if actualValue != v {
			t.Errorf("Expecting %v=%v, got %v\n", k, v, actualValue)
		}
	}
}

func TestParseCommandParametersMixedModeEdges(t *testing.T) {
	parser, _ := NewPseudoJsonParser()
	testMessage := `{ x: 1} y: 2 `

	_, err := ParseCommandParameters(parser, testMessage)
	if err != nil {
		t.Errorf("unable to parse message: %v: %v\n", testMessage, err)
	}
}

func TestForParseErrors(t *testing.T) {
	testMessages := []string{
		`{ kala: "maja" }`,
		`{ kala: "maja", puu: "juur" }`,
		`{ kala: "maja", int: 1234, float: 12.34 }`,
		`{ driver: { name: "PyMongo", version: "3.4.0" }, os: { type: "Linux" } }`,
		`{ kala: ObjectId('5a2fc7bd9b45c7117bee26c5') }`,
		`{ $kala: "maja" }`,
		`{ count: "mycatpicscollection", query: { MyObjectId: ObjectId('5a2fc7bd9b45c7117bee26c5'),
		   baz.max_time: { $gte: 1523022862.698 }, baz.min_time: { $lte: 1523022882.698 },
		   baz.category: "catinabag" }, $readPreference: { mode: "secondaryPreferred" }, $db: "FooDb" }`,
		// This next one is interesting because it contains a bracketed list
		`{ find: "mycatpicscollection", filter: { foo.FooObjectId: ObjectId('5a8c3a142053a407a936745e'),
		   foo.max_time: { $gte: 1534769530.5 }, foo.min_time: { $lte: 1534769548.47 },
		   foo.category: { $in: [ "alley", "home" ] } }, $db: "FooDb" }`,
		// And this is interesting because it contains all sorts of shit that is outside
		// of curly braces. Should we attempt to support this or parse out with regex?
		`{ foo.FooObjectId: 1, foo.category: 1, foo.min_time: -1, foo.max_time: 1 }
		   keysExamined:50314 docsExamined:2 cursorExhausted:1 numYields:393 nreturned:2 reslen:14980
		   locks:{ Global: { acquireCount: { r: 788 } }, Database: { acquireCount: { r: 394 } },
		   Collection: { acquireCount: { r: 394 } } }`,
	}
	parser, err := NewPseudoJsonParser()
	if err != nil {
		t.Errorf("Error initializing parser: %v\n", err)
	}

	for _, v := range testMessages {
		if _, err := ParseCommandParameters(parser, v); err != nil {
			t.Errorf("unable to parse message: %v: %v\n", v, err)
		}
	}
}

func benchmarkParseCommandParameters(testMessage string, b *testing.B) {
	parser, err := NewPseudoJsonParser()
	if err != nil {
		b.Errorf("Error initializing parser: %v\n", err)
	}

	for i := 0; i < b.N; i++ {
		if _, err := ParseCommandParameters(parser, testMessage); err != nil {
			b.Errorf("Cannot parse: %v\n", testMessage)
		}
	}
}

func BenchmarkParseCommandParametersSmall(b *testing.B) {
	benchmarkParseCommandParameters(`{ a: 1 }`, b)
}

func BenchmarkParseCommandParametersMedium(b *testing.B) {
	benchmarkParseCommandParameters(`{ driver: { name: "PyMongo", version: "3.4.0" }, os: { type: "Linux" } }`, b)
}

func BenchmarkParseCommandParametersLarge(b *testing.B) {
	benchmarkParseCommandParameters(`{ 
	  count: "mycatpicscollection", query: { MyObjectId: ObjectId('5a2fc7bd9b45c7117bee26c5'),
	  baz.max_time: { $gte: 1523022862.698 }, baz.min_time: { $lte: 1523022882.698 },
	  baz.category: "catinabag" }, $readPreference: { mode: "secondaryPreferred" }, $db: "FooDb"}`, b)
}

// Compare against a regex based parser
func TestParseCommandParametersRegex(t *testing.T) {
	testMessage := `{ driver: { name: "PyMongo", version: "3.4.0" }, os: { type: "Linux" } }`
	re := regexp.MustCompile(`{ driver: { name: "(?P<driverName>.*)", version: "(?P<driverVersion>.*)" }, os: { type: "(?P<osType>.*)" } }`)
	match := re.FindStringSubmatch(testMessage)
	matches := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 {
			matches[name] = match[i]
		}
	}

	var expectMatches = map[string]string{
		"driverName":    "PyMongo",
		"driverVersion": "3.4.0",
		"osType":        "Linux",
	}

	for k, v := range expectMatches {
		if matches[k] != v {
			t.Errorf("Expecting %v=%v, got %v\n", k, v, matches[k])
		}
	}
}

// Compare against a regex based parser
func BenchmarkParseCommandParametersRegexMedium(b *testing.B) {
	testMessage := `{ driver: { name: "PyMongo", version: "3.4.0" }, os: { type: "Linux" } }`
	re := regexp.MustCompile(`{ driver: { name: "(?P<driverName>.*)", version: "(?P<driverVersion>.*)" }, os: { type: "(?P<osType>.*)" } }`)

	for i := 0; i < b.N; i++ {
		_ = re.FindStringSubmatch(testMessage)
	}
}
