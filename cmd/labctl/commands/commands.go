package commands

type Command string

const (
	Start  Command = "start"
	Delete Command = "delete"
	List   Command = "list"
)

func Parse(value string) (Command, bool) {
	command := Command(value)
	switch command {
	case Start, Delete, List:
		return command, true
	default:
		return "", false
	}
}
