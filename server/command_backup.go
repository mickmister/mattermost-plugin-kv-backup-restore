package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeBackup(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	keys, appErr := p.API.KVList(0, 10000)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error listing keys. err=%v", appErr)
	}

	if len(keys) == 0 || keys[0] == "null" {
		return p.responsef(cmdArgs, "No keys found.")
	}

	allValues := ""
	for i, key := range keys {
		value, appErr := p.API.KVGet(key)
		if appErr != nil {
			return p.responsef(cmdArgs, "Error getting value for key `%s`. err=%v", key, appErr)
		}

		s := `"` + string(value) + `"`
		if len(s) > 2 && (s[1] == '{' || s[1] == '[') {
			var buf bytes.Buffer
			err := json.Indent(&buf, value, "  ", "  ")
			if err == nil {
				s = string(buf.Bytes())
			}
		}

		comma := ",\n"
		if i == len(keys)-1 {
			comma = ""
		}
		allValues += fmt.Sprintf("  \"%s\": %s%s", key, s, comma)
	}

	asJSON := fmt.Sprintf("{\n%s\n}", allValues)

	post := &model.Post{
		UserId:    cmdArgs.UserId,
		ChannelId: cmdArgs.ChannelId,
	}

	res := fmt.Sprintf("```json\n%s\n```", asJSON)
	if len(args) != 0 && args[0] == "file" {
		fileInfo, appErr := p.API.UploadFile([]byte(asJSON), cmdArgs.ChannelId, manifest.Id+"-backup.json")
		if appErr != nil {
			return p.responsef(cmdArgs, "Error uploading result err=%v", appErr)
		}

		post.FileIds = append(post.FileIds, fileInfo.Id)
		res = fmt.Sprintf("Backed up %d values", len(keys))
	}

	post.Message = res
	p.API.CreatePost(post)
	return &model.CommandResponse{}
}
