package constants

// Component names for error tracking
const (
	ComponentBronze = "Bronze"
	ComponentSilver = "Silver"
	ComponentGold   = "Gold"
)

// Default paths
const (
	DefaultConfigPath = "config/config.yaml"
)

// Error messages
const (
	ErrMsgConfigLoad     = "failed to load configuration"
	ErrMsgDBConnection   = "failed to connect to database"
	ErrMsgQueryExecution = "failed to execute query"
	ErrMsgDataExtraction = "failed to extract data"
	ErrMsgDataTransform  = "failed to transform data"
	ErrMsgAIProcessing   = "failed to process with AI"
	ErrMsgFileOperation  = "failed to perform file operation"
)
