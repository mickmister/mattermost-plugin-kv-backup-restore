package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeShow(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a key.")
	}

	key := args[0]
	value, appErr := p.API.KVGet(key)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error getting value. err=%v", appErr)
	}

	if len(value) == 0 || string(value) == "null" {
		return p.responsef(cmdArgs, "Key `%s` not found", key)
	}

	res := renderValue(value)
	return p.responsef(cmdArgs, res)
}
