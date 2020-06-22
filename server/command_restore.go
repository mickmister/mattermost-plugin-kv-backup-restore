package main

import (
	"encoding/json"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeRestore(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide data to restore.")
	}

	out := map[string][]byte{}
	values := map[string]interface{}{}
	var data []byte

	if args[0] == "file" {
		if len(args) == 1 {
			return p.responsef(cmdArgs, "Please provide a file id.")
		}

		fileID := args[1]
		file, appErr := p.API.GetFile(fileID)
		if appErr != nil {
			return p.responsef(cmdArgs, "Error fetching file `%s`. err=%v", fileID, appErr)
		}

		data = file
	} else {
		data = []byte(strings.Join(args, " "))
	}

	err := json.Unmarshal(data, &values)
	if err != nil {
		return p.responsef(cmdArgs, "Error unmarshaling payload. err=%v", err)
	}

	for key, value := range values {
		var toSave []byte

		switch value.(type) {
		case string:
			toSave = []byte(value.(string))
		default:
			b, err := json.Marshal(value)
			if err != nil {
				return p.responsef(cmdArgs, "Error unmarshaling key `%s`'s value. err=%v", key, err)
			}
			toSave = b
		}

		out[key] = toSave
	}

	appErr := p.API.KVDeleteAll()
	if appErr != nil {
		return p.responsef(cmdArgs, "Failed to clear the kv store.")
	}

	for key, value := range out {
		appErr := p.API.KVSet(key, value)
		if appErr != nil {
			return p.responsef(cmdArgs, "Error setting key `%s`'s value. err=%v", key, err)
		}
	}

	return executeBackup(p, c, cmdArgs)
}
