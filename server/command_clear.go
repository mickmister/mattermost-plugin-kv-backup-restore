package main

import (
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeClear(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	keys, appErr := p.API.KVList(0, 1)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error listing keys. err=%v", appErr)
	}

	if len(keys) == 0 || keys[0] == "null" {
		return p.responsef(cmdArgs, "No keys found.")
	}

	appErr = p.API.KVDeleteAll()
	if appErr != nil {
		return p.responsef(cmdArgs, "Error clearing keys. err=%v", appErr)
	}

	return p.responsef(cmdArgs, "Successfully cleared the kvstore")
}
