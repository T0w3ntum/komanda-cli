package command

import "github.com/mephux/komanda-cli/client"

var (
	Commands       []Command
	Server         *client.Server
	CurrentChannel = client.StatusChannel
)

type MetadataTmpl struct {
	name        string
	description string
	aliases     []string
}

func (c *MetadataTmpl) Name() string {
	return c.name
}

func (c *MetadataTmpl) Description() string {
	return c.description
}

func (c *MetadataTmpl) Aliases() []string {
	return c.aliases
}

type CommandMetadata interface {
	Name() string
	Description() string
	Aliases() []string
}

type Command interface {
	Metadata() CommandMetadata
	Exec(args []string) error
}

func Register(server *client.Server) {
	Server = server

	Commands = []Command{
		exitCmd(),
		connectCmd(),
		statusCmd(),
		helpCmd(),
		joinCmd(),
		partCmd(),
		clearCmd(),
		logoCmd(),
		versionCmd(),
		nickCmd(),
		passCmd(),
		rawCmd(),
	}
}

func Get(cmd string) Command {

	for _, command := range Commands {
		metadata := command.Metadata()

		if metadata.Name() == cmd {
			return command
		} else {
			for _, c := range metadata.Aliases() {
				if c == cmd {
					return command
				}
			}
		}
	}

	return emptyCmd()
}

func Run(command string, args []string) error {
	return Get(command).Exec(args)
}
