package mongolog

import (
	_ "fmt"
	"github.com/alecthomas/participle"
)

type ElementMap map[string]*Value

type PseudoJson struct {
	// This mess is here to support the way Mongo intermixes key-value pairs with and without
	// the braces. Not sure that I've nailed it, but it appears to support most of it.
	// Though maybe just use `"{" @@ { "," @@ } "}"` and deal regex the rest of it.
	// TODO: Actually moving the messy parts to another parser so clean this up when done
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
	StringValue   string      `( (@String|"true"|"false"|"COLLSCAN"|"IXSCAN")`
	NumericValue  float64     `| @(["-"] (Int | Float))`
	ObjectIdValue string      `| Ident "(" @String ")"`
	ArrayValue    []*Value    `| "[" { @@ { "," @@ } } "]"`
	Nested        *PseudoJson `| @@ )`
}

type PseudoJsonParser struct {
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

func NewPseudoJsonParser() (parser PseudoJsonParser, err error) {
	parser = PseudoJsonParser{}
	parser.p, err = participle.Build(&PseudoJson{})
	return
}

func NewPlanSummaryParser() (parser PseudoJsonParser, err error) {
	parser = PseudoJsonParser{}
	parser.p, err = participle.Build(&PlanSummary{})
	return
}

func ParseMessage(parser PseudoJsonParser, message string) (*PseudoJson, error) {
	mongoJson := &PseudoJson{}
	if err := parser.p.ParseString(message, mongoJson); err != nil {
		return mongoJson, err
	}

	mapElementKeys(mongoJson)

	return mongoJson, nil
}

func ParsePlanSummary(parser PseudoJsonParser, message string) (result *PlanSummary, err error) {
	result = &PlanSummary{}
	err = parser.p.ParseString(message, result)
	return
}
