package event

type EventHandler string

var (
	CliHandler    EventHandler = "cli"
	SqsHandler    EventHandler = "sqs"
	AssumeRoleArn string
)

type Event interface {
	Init(cfg string) error
	GetUri() (string, error)
}

func New(handler EventHandler) Event {
	switch handler {
	case CliHandler:
		return &Cli{}
	case SqsHandler:
		return &Sqs{}
	default:
		return nil
	}
}
