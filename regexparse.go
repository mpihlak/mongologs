package mongolog

import (
	"regexp"
)

var (
	MongoLoglineRegex = regexp.MustCompile(
		`(?P<timestamp>[^\s]+)\s` +
			`(?P<severity>.)\s` +
			`(?P<component>[^\s]+)\s+` +
			`(?P<context>[^\s]+)\s` +
			`(?P<message>.*)`)

	MongoLogPayloadRegex = regexp.MustCompile(
		`command (?P<collection>[^\s]+)\scommand:\s` +
			`(?P<command>[^\s]+)\s` +
			`(?P<commandparams>{.*})\s` +
			`planSummary:\s` +
			`(?P<plansummary>.*)\sprotocol:` +
			`(?P<protocol>[^\s]+)\s` +
			`(?P<duration>[0-9]+)ms`)
)

// Match a regexp against a string. Return the subgroups in a dictionary
func RegexpMatch(re *regexp.Regexp, matchText string) (results map[string]string) {
	match := re.FindStringSubmatch(matchText)
	if match == nil {
		return
	}

	results = make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i != 0 && i < len(match) {
			results[name] = match[i]
		}
	}
	return
}
