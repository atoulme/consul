package command

import (
	"fmt"

	"github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/consul/configutil"
)

// ValidateCommand is a Command implementation that is used to
// verify config files
type ValidateCommand struct {
	BaseCommand

	// flags
	configFiles []string
	quiet       bool
}

func (c *ValidateCommand) initFlags() {
	c.InitFlagSet()
	c.FlagSet.Var((*configutil.AppendSliceValue)(&c.configFiles), "config-file",
		"Path to a JSON file to read configuration from. This can be specified multiple times.")
	c.FlagSet.Var((*configutil.AppendSliceValue)(&c.configFiles), "config-dir",
		"Path to a directory to read configuration files from. This will read every file ending in "+
			".json as configuration in this directory in alphabetical order.")
	c.FlagSet.BoolVar(&c.quiet, "quiet", false,
		"When given, a successful run will produce no output.")
	c.HideFlags("config-file", "config-dir")
}

func (c *ValidateCommand) Help() string {
	c.initFlags()
	return c.HelpCommand(`
Usage: consul validate [options] FILE_OR_DIRECTORY...

  Performs a basic sanity test on Consul configuration files. For each file
  or directory given, the validate command will attempt to parse the
  contents just as the "consul agent" command would, and catch any errors.
  This is useful to do a test of the configuration only, without actually
  starting the agent.

  Returns 0 if the configuration is valid, or 1 if there are problems.

`)
}

func (c *ValidateCommand) Run(args []string) int {
	c.initFlags()
	if err := c.FlagSet.Parse(args); err != nil {
		c.UI.Error(err.Error())
		return 1
	}

	if len(c.FlagSet.Args()) > 0 {
		c.configFiles = append(c.configFiles, c.FlagSet.Args()...)
	}

	if len(c.configFiles) < 1 {
		c.UI.Error("Must specify at least one config file or directory")
		return 1
	}

	b, err := config.NewBuilder(config.Flags{ConfigFiles: c.configFiles})
	if err != nil {
		c.UI.Error(fmt.Sprintf("Config validation failed: %v", err.Error()))
		return 1
	}
	if _, err := b.BuildAndValidate(); err != nil {
		c.UI.Error(fmt.Sprintf("Config validation failed: %v", err.Error()))
		return 1
	}

	if !c.quiet {
		c.UI.Output("Configuration is valid!")
	}
	return 0
}

func (c *ValidateCommand) Synopsis() string {
	return "Validate config files/directories"
}
