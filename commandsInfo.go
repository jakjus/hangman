package main

import (
	"github.com/bwmarrin/discordgo"
)

var (
	integerOptionMinValue          = 1.0
	dmPermission                   = false
	defaultMemberPermissions int64 = discordgo.PermissionManageServer

	commands = []*discordgo.ApplicationCommand{
		{
			Name: "play",
			// All commands and options must have a description
			// Commands/options without description will fail the registration
			// of the command.
			Description: "Start Hangman Game",
		},
	}
)
