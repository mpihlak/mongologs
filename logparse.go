package mongolog

import (
	_ "fmt"
	"github.com/alecthomas/participle"
)

type ElementMap map[string]*KeyValue

type PseudoJson struct {
	Elements []*KeyValue `"{" @@ { "," @@ } "}"`

	// Key the elements by KeyValue for convenient access
	elems ElementMap
}

type KeyValue struct {
	// Key is an identifier that maybe has dots and maybe starts with a $ sign.
	Key string `@((Ident | "$" Ident) { "." Ident }) ":"`

	StringValue   string      `( @String`
	NumericValue  float64     `| @(["-"] (Int | Float))`
	ObjectIdValue string      `| "ObjectId" "(" @String ")"`
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
		mongoJson.elems[e.Key] = e
		if e.Nested != nil {
			mapElementKeys(e.Nested)
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
