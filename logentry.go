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

	parser.connectionMap = make(map[string]*Connection)

	return
}

func handleNewConnection(parser LogParser, entry MongoLogEntry, connParams map[string]string) {
	conn := &Connection{
		ConnectionId: "[conn" + connParams["id"] + "]",
		IpAddress:    connParams["ip"],
		Port:         connParams["port"],
	}
	parser.connectionMap[conn.ConnectionId] = conn
}

func handleCloseConnection(parser LogParser, entry MongoLogEntry) {
	delete(parser.connectionMap, entry.Context)
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
				handleNewConnection(result, connParams)
			}
		} else {
			connParams := RegexpMatch(MongoConnectionMetadataRegex, result.LogMessage)
			if connParams != nil {
				handleConnectionMetadata(result, connParams)
			} else {
				connParams := RegexpMatch(MongoEndConnectionRegex, result.LogMessage)
				if connParams != nil {
					handleCloseConnection(result, connParams)
				}
			}


				conn, ok := parser.connectionMap[result.Context]
				if !ok {
					conn = &Connection{
						ConnectionId: result.Context,
						IpAddress:    connParams["ip"],
						Port:         connParams["port"],
					}
					parser.connectionMap[result.Context] = conn
				}
				// TODO: parse and add the metadata
			}

			// TODO: try also endconnection and clean up stale entries from the map
		}
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
