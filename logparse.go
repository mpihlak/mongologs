package mongolog

import (
	_ "fmt"
	"github.com/alecthomas/participle"
)

type ElementMap map[string]*Value

type PseudoJson struct {
	// This mess is here to support the way Mongo intermixes key-value pairs with and without
	// the braces, sometimes separated by commas and somtimes not. Not sure that I've nailed it, but
	// it appears to support most of it.  Though maybe just use `"{" @@ { "," @@ } "}"` and regex
	// the rest of it.
	Elements []*KeyValue `(@@  { (@@ | "{" @@ { "," @@ } "}") }) | (("{" { @@ { ","  @@ } } "}" ) { @@ })`

	// Key the elements by KeyValue for convenient access
	elems ElementMap
}

type PlanSummary struct {
	Items []*PlanItem `@@ { "," @@ }`
}

type PlanItem struct {
	PlanType string      `@Ident` // Like IXSCAN or COLSCAN
	PlanInfo *PseudoJson `@@`
}

type KeyValue struct {
	// Key is an identifier that maybe has dots and maybe starts with a $ sign.
	Key string `@((Ident | "$" Ident) { "." Ident }) ":"`
	Val *Value `@@`
}

type Value struct {
	StringValue  string         `( (@String|"true"|"false"|"null")`
	NumericValue float64        `| @(["-"] (Int | Float))`
	FuncValue    *FunctionValue `| @@`
	ArrayValue   []*Value       `| "[" { @@ { "," @@ } } "]"`
	Nested       *PseudoJson    `| @@ )`
}

type FunctionValue struct {
	//FuncName string   `["new"] @Ident`
	FuncName string   `["new"] @("ObjectId"|"UUID"|"BinData"|"Date"|"Timestamp")`
	FuncArgs []*Value `"(" { @@ ({ "," @@ }) } ")"`
}

type MongoLogParser struct {
	p *participle.Parser
}

func mapElementKeys(mongoJson *PseudoJson) {
	mongoJson.elems = make(ElementMap)
	for _, e := range mongoJson.Elements {
		mongoJson.elems[e.Key] = e.Val
		if e.Val.Nested != nil {
			mapElementKeys(e.Val.Nested)
		}
	}
}

func NewPseudoJsonParser() (parser MongoLogParser, err error) {
	parser = MongoLogParser{}
	parser.p, err = participle.Build(&PseudoJson{})
	return
}

func NewCommandParametersParser() (parser MongoLogParser, err error) {
	return NewPseudoJsonParser()
}

func NewPlanSummaryParser() (parser MongoLogParser, err error) {
	parser = MongoLogParser{}
	parser.p, err = participle.Build(&PlanSummary{})
	return
}

func ParsePseudoJson(parser MongoLogParser, message string) (result *PseudoJson, err error) {
	result = &PseudoJson{}
	err = parser.p.ParseString(message, result)
	if err == nil {
		mapElementKeys(result)
	}

	return
}

func ParseCommandParameters(parser MongoLogParser, message string) (result *PseudoJson, err error) {
	return ParsePseudoJson(parser, message)
}

func ParsePlanSummary(parser MongoLogParser, message string) (result *PlanSummary, err error) {
	result = &PlanSummary{}
	err = parser.p.ParseString(message, result)
	return
}
