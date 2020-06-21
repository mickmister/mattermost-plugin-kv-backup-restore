package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeDelete(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a key.")
	}

	key := args[0]
	value, appErr := p.API.KVGet(key)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error getting value for key `%s`. err=%v", key, appErr)
	}

	if len(value) == 0 {
		return p.responsef(cmdArgs, "No value found for key `%s", key)
	}

	appErr = p.API.KVDelete(key)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error deleting key `%s`", key)
	}

	return p.responsef(cmdArgs, "Key `%s` successfully deleted", key)
}
