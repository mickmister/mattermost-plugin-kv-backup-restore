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

	key := args[0]

	stored, appErr := p.API.KVGet(key)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error getting key `%s`'s value. err=%v", key, appErr)
	}

	res := fmt.Sprintf("Updated key: `%s`", key)
	if len(stored) == 0 {
		res = fmt.Sprintf("New key added: `%s`", key)
	}

	var data []byte
	if args[1] == "file" {
		if len(args) == 2 && args[1] == "file" {
			return p.responsef(cmdArgs, "Please provide a file id.")
		}

		fileID := args[2]
		file, appErr := p.API.GetFile(fileID)
		if appErr != nil {
			return p.responsef(cmdArgs, "Error fetching file `%s`. err=%v", fileID, appErr)
		}

		data = file
	} else {
		data = []byte(strings.Join(args[1:], " "))
	}

	appErr = p.API.KVSet(key, data)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error setting value. err=%v", appErr)
	}

	s := renderValue(data)
	res += s
	return p.responsef(cmdArgs, res)
}
