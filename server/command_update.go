package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeUpdate(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a key.")
	}

	if len(args) == 1 {
		return p.responsef(cmdArgs, "Please provide a value to set.")
	}

	key, valueFull := args[0], args[1:]
	value := strings.Join(valueFull, " ")

	stored, appErr := p.API.KVGet(key)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error getting key `%s`'s value. err=%v", key, appErr)
	}

	res := fmt.Sprintf("Updated key: `%s`", key)
	if len(stored) == 0 {
		res = fmt.Sprintf("New key added: `%s`", key)
	}

	appErr = p.API.KVSet(key, []byte(value))
	if appErr != nil {
		return p.responsef(cmdArgs, "Error setting value. err=%v", appErr)
	}

	s := renderValue([]byte(value))
	res += s
	return p.responsef(cmdArgs, res)
}
