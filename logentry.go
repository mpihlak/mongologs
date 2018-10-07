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
	ConnectionInfo    *Connection
	CommandParameters *PseudoJson
	PlanInfo          *PlanSummary
}

type Connection struct {
	ConnectionId string
	IpAddress    string
	Port         string
}

type LogParser struct {
	commandParametersParser MongoLogParser
	planSummaryParser       MongoLogParser
	connectionMetaParser	MongoLogParser

	connectionMap map[string]*Connection
}

func NewLogParser() (parser LogParser, err error) {
	parser.commandParametersParser, err = NewCommandParametersParser()
	if err != nil {
		return parser, fmt.Errorf("Cannot initialize commandParametersParser: %v", err)
	}

	parser.planSummaryParser, err = NewPlanSummaryParser()
	if err != nil {
		return parser, fmt.Errorf("Cannot initialize planSummaryParser: %v", err)
	}

	parser.connectionMetaParser, err = NewCommandParametersParser()
	if err != nil {
		return parser, fmt.Errorf("Cannot initialize connectionMetaParser: %v", err)
	}

	parser.connectionMap = make(map[string]*Connection)
	return
}

func handleNewConnection(parser LogParser, entry *MongoLogEntry, connParams map[string]string) {
	conn := &Connection{
		ConnectionId: "[conn" + connParams["id"] + "]",
		IpAddress:    connParams["ip"],
		Port:         connParams["port"],
	}
	parser.connectionMap[conn.ConnectionId] = conn
	entry.ConnectionInfo = conn
}

func handleCloseConnection(parser LogParser, entry *MongoLogEntry) {
	if conn, ok := parser.connectionMap[entry.Context]; ok {
		entry.ConnectionInfo = conn
		delete(parser.connectionMap, entry.Context)
	}
}

func handleConnectionMetadata(parser LogParser, entry MongoLogEntry, connParams map[string]string) {
	if conn, ok := parser.connectionMap[entry.Context]; ok {
		// Parse the metadata payload and add to connection
		connMeta, err := ParsePseudoJson(parser.connectionMetaParser, connParams["metadata"])
		if err == nil {
			_  = conn
			_ = connMeta
		}
	}
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

	if result.Component == "NETWORK" {
		if result.Context == "[listener]" {
			connParams := RegexpMatch(MongoNewConnectionRegex, result.LogMessage)
			if connParams != nil {
				handleNewConnection(parser, &result, connParams)
			}
		} else {
			connParams := RegexpMatch(MongoConnectionMetadataRegex, result.LogMessage)
			if connParams != nil {
				handleConnectionMetadata(parser, result, connParams)
			} else {
				connParams := RegexpMatch(MongoEndConnectionRegex, result.LogMessage)
				if connParams != nil {
					handleCloseConnection(parser, &result)
				}
			}
		}
	}

	if conn, ok := parser.connectionMap[result.Context]; ok {
		result.ConnectionInfo = conn
	}

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
