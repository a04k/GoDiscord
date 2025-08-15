package roles

import (
	"DiscordBot/commands"
)

func init() {
	module := &commands.ModuleInfo{
		Name:         "Roles",
		Description:  "Role management commands for creating and assigning roles",
		Version:      "1.0.0",
		Author:       "Bot Team",
		Category:     "Roles",
		Dependencies: []string{},
		Config: map[string]interface{}{
			"max_roles_per_user":  10,
			"allow_role_creation": true,
		},
		Commands: []commands.CommandInfo{
			{
				Name:        "createrole",
				Aliases:     []string{},
				Description: "Creates a new role",
				Usage:       ".createrole <name>",
				Category:    "Roles",
			},
			{
				Name:        "setrole",
				Aliases:     []string{},
				Description: "Assigns a role to a user",
				Usage:       ".setrole <user> <role>",
				Category:    "Roles",
			},
			{
				Name:        "removerole",
				Aliases:     []string{"rr"},
				Description: "Removes a role from a user",
				Usage:       ".removerole <user> <role>",
				Category:    "Roles",
			},
			{
				Name:        "inrole",
				Aliases:     []string{},
				Description: "Lists users in a role",
				Usage:       ".inrole <role>",
				Category:    "Roles",
			},
			{
				Name:        "roleinfo",
				Aliases:     []string{},
				Description: "Shows information about a role",
				Usage:       ".roleinfo <role>",
				Category:    "Roles",
			},
		},
		SlashCommands: []commands.SlashCommandInfo{},
	}

	commands.RegisterModule(module)

	// Register command handlers
	commands.RegisterCommand("createrole", CreateRole)
	commands.RegisterCommand("setrole", SetRole)
	commands.RegisterCommand("removerole", RemoveRole, "rr")
	commands.RegisterCommand("inrole", InRole)
	commands.RegisterCommand("roleinfo", RoleInfo)
}
