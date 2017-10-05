package command

import (
	"flag"
	"fmt"
)

type OperatorRaftCommand struct {
	BaseCommand

	// flags
	listPeers  bool
	removePeer bool
	address    string
}

func (c *OperatorRaftCommand) initFlags() {
	c.InitFlagSet()
	// todo(fs): should we remove these flags according to the comment below?
	c.FlagSet.BoolVar(&c.listPeers, "list-peers", false,
		"If this flag is provided, the current Raft peer configuration will be "+
			"displayed. If the cluster is in an outage state without a leader, you may need "+
			"to set -stale to 'true' to get the configuration from a non-leader server.")
	c.FlagSet.BoolVar(&c.removePeer, "remove-peer", false,
		"If this flag is provided, the Consul server with the given -address will be "+
			"removed from the Raft configuration.")

	c.FlagSet.StringVar(&c.address, "address", "",
		"The address to remove from the Raft configuration.")

	// Leave these flags for backwards compatibility, but hide them
	// TODO: remove flags/behavior from this command in Consul 0.9
	c.HideFlags("list-peers", "remove-peer", "address")
}

func (c *OperatorRaftCommand) Help() string {
	c.initFlags()
	return c.HelpCommand(`
Usage: consul operator raft <subcommand> [options]

The Raft operator command is used to interact with Consul's Raft subsystem. The
command can be used to verify Raft peers or in rare cases to recover quorum by
removing invalid peers.

Subcommands:

    list-peers     Display the current Raft peer configuration
    remove-peer    Remove a Consul server from the Raft configuration

`)
}

func (c *OperatorRaftCommand) Synopsis() string {
	return "Provides cluster-level tools for Consul operators"
}

func (c *OperatorRaftCommand) Run(args []string) int {
	if result := c.raft(args); result != nil {
		c.UI.Error(result.Error())
		return 1
	}
	return 0
}

// raft handles the raft subcommands.
func (c *OperatorRaftCommand) raft(args []string) error {
	c.initFlags()
	if err := c.FlagSet.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil
		}
		return err
	}

	// Set up a client.
	client, err := c.HTTPClient()
	if err != nil {
		return fmt.Errorf("error connecting to Consul agent: %s", err)
	}

	// Dispatch based on the verb argument.
	if c.listPeers {
		result, err := raftListPeers(client, c.HTTPStale())
		if err != nil {
			c.UI.Error(fmt.Sprintf("Error getting peers: %v", err))
		}
		c.UI.Output(result)
	} else if c.removePeer {
		if err := raftRemovePeers(c.address, "", client.Operator()); err != nil {
			return fmt.Errorf("Error removing peer: %v", err)
		}
		c.UI.Output(fmt.Sprintf("Removed peer with address %q", c.address))
	} else {
		c.UI.Output(c.Help())
		return nil
	}

	return nil
}
