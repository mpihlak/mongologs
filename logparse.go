package mongolog

import (
	_ "fmt"
	"github.com/alecthomas/participle"
)

type ElementMap map[string]*Value

type PseudoJson struct {
	// Elements []*KeyValue `"{" @@ { "," @@ } "}"`
	Elements []*KeyValue `(@@  { (@@ | "{" @@ "}") }) | (("{" @@ { ","  @@ } "}" ) { @@ })`

	// Key the elements by KeyValue for convenient access
	elems ElementMap
}

type KeyValue struct {
	// Key is an identifier that maybe has dots and maybe starts with a $ sign.
	Key string `@((Ident | "$" Ident) { "." Ident }) ":"`
	Val *Value `@@`
}

type Value struct {
	StringValue   string      `( @String`
	NumericValue  float64     `| @(["-"] (Int | Float))`
	ObjectIdValue string      `| "ObjectId" "(" @String ")"`
	ArrayValue    []*Value    `| "[" @@ { "," @@ } "]"`
	Nested        *PseudoJson `| @@ )`
}

type PseudoJsonParser struct {
	p *participle.Parser
}

func NewPseudoJsonParser() (PseudoJsonParser, error) {
	parser := PseudoJsonParser{}
	var err error
	parser.p, err = participle.Build(&PseudoJson{})
	return parser, err
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

func ParseMessage(parser PseudoJsonParser, message string) (*PseudoJson, error) {
	mongoJson := &PseudoJson{}
	if err := parser.p.ParseString(message, mongoJson); err != nil {
		return mongoJson, err
	}

	mapElementKeys(mongoJson)

	return mongoJson, nil
}
