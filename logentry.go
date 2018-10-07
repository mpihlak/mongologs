package mongolog

import (
	"fmt"
)

type MongoLogEntry struct {
	Timestamp         string
	Severity          string
	Component         string
	Context           string
	LogMessage        string
	CommandParameters *PseudoJson
	PlanInfo          *PlanSummary
}

type LogParser struct {
	commandParametersParser MongoLogParser
	planSummaryParser       MongoLogParser
}

func NewLogParser() (parser LogParser, err error) {
	parser.commandParametersParser, err = NewPseudoJsonParser()
	if err != nil {
		return parser, fmt.Errorf("Cannot initialize commandParametersParser: %v", err)
	}
	parser.planSummaryParser, err = NewPlanSummaryParser()
	if err != nil {
		return parser, fmt.Errorf("Cannot initialize planSummaryParser: %v", err)
	}

	return
}

// ParseLogEntry parses the MongoDb log line into MongoLogEntry structure
func ParseLogEntry(parser LogParser, logLine string) (result MongoLogEntry, err error) {
	logMatch := RegexpMatch(MongoLoglineRegex, logLine)
	if logMatch == nil {
		return result, fmt.Errorf("logLine does not match Mongo log pattern.")
	}
	result.Timestamp = logMatch["timestamp"]
	result.Severity = logMatch["severity"]
	result.Component = logMatch["component"]
	result.Context = logMatch["context"]
	result.LogMessage = logMatch["message"]

	if result.Severity == "I" && result.Component == "COMMAND" {
		// Parse the command parameters and execution plan
		commandBody := RegexpMatch(MongoLogPayloadRegex, result.LogMessage)
		if commandBody == nil {
			return result, fmt.Errorf("COMMAND payload does not match expected.")
		}

		result.CommandParameters, err = ParseCommandParameters(parser.commandParametersParser,
			commandBody["commandparams"])
		if err != nil {
			return result, fmt.Errorf("commandparams: parse error: %v", err)
		}

		result.PlanInfo, err = ParsePlanSummary(parser.planSummaryParser,
			commandBody["plansummary"])
		if err != nil {
			return result, fmt.Errorf("plansummary: pare error: %v", err)
		}
	}

	return
}
