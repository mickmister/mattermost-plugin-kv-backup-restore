package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

func executeRestoreSQL(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please a plugin ID")
	}

	out := map[string]*model.PluginKeyValue{}
	valuesWithString := map[string]*KVWithString{}
	var data []byte

	var fileID string
	var err error
	if len(args) == 1 {
		fileID, err = p.getRecentPostFileID(cmdArgs.ChannelId)
		if err != nil {
			return p.responsef(cmdArgs, "Error getting file id from previous post. err=%v", err)
		}
	} else {
		fileID = args[1]
	}

	file, appErr := p.API.GetFile(fileID)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error fetching file `%s`. err=%v", fileID, appErr)
	}

	data = file
	err = json.Unmarshal(data, &valuesWithString)
	if err != nil {
		return p.responsef(cmdArgs, "Error unmarshaling payload. err=%v", err)
	}

	for key, value := range valuesWithString {
		var toSave []byte

		toSave, err = base64.StdEncoding.DecodeString(value.Value)
		if err != nil {
			return p.responsef(cmdArgs, "Error decoding key `%s`. err=%v", key, err)
		}

		out[key] = &model.PluginKeyValue{
			PluginId: value.PluginId,
			Key:      value.Key,
			Value:    toSave,
			ExpireAt: value.ExpireAt,
		}
	}

	err = p.deleteAllKeys(args[0])
	if err != nil {
		return p.responsef(cmdArgs, "Error deleting keys. err=%v", err)
	}

	db, err := p.client.Store.GetMasterDB()
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "failed to get a database connection").Error())
	}

	for key, value := range out {
		_, err := db.Exec("INSERT into PluginKeyValueStore VALUES (?, ?, ?, ?)", value.PluginId, value.Key, value.Value, value.ExpireAt)
		if err != nil {
			return p.responsef(cmdArgs, errors.Wrap(err, "failed to insert for key "+key).Error())
		}
	}

	return p.responsef(cmdArgs, "Successfully restored %d values", len(out))
}

func (p *Plugin) insertKeyValue(pluginID, key string, valye []byte) error {
	db, err := p.client.Store.GetMasterDB()
	if err != nil {
		return errors.Wrap(err, "failed to get a database connection")
	}

	q := fmt.Sprintf(`
		DELETE FROM PluginKeyValueStore WHERE PluginId = '%s'
	`, pluginID)

	_, err = db.Query(q)
	if err != nil {
		return errors.Wrap(err, "error querying database")
	}

	return nil
}

func (p *Plugin) deleteAllKeys(pluginID string) error {
	db, err := p.client.Store.GetMasterDB()
	if err != nil {
		return errors.Wrap(err, "failed to get a database connection")
	}

	q := fmt.Sprintf(`
		DELETE FROM PluginKeyValueStore WHERE PluginId = '%s'
	`, pluginID)

	_, err = db.Query(q)
	if err != nil {
		return errors.Wrap(err, "error querying database")
	}

	return nil
}

func executeRestore(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide data to restore.")
	}

	out := map[string][]byte{}
	values := map[string]interface{}{}
	var data []byte

	if args[0] == "file" {
		var fileID string
		var err error
		if len(args) == 1 {
			fileID, err = p.getRecentPostFileID(cmdArgs.ChannelId)
			if err != nil {
				return p.responsef(cmdArgs, "Error getting file id from previous post. err=%v", err)
			}
		} else {
			fileID = args[1]
		}

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
			if isGeneratedKeyValue(key) || true {
				toSave, err = base64.StdEncoding.DecodeString(value.(string))
				if err != nil {
					return p.responsef(cmdArgs, "Error decoding key `%s`. err=%v", key, err)
				}
			}
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

	return p.responsef(cmdArgs, "Successfully restored %d values", len(out))
}

func (p *Plugin) getRecentPostFileID(channelID string) (string, error) {
	posts, appErr := p.API.GetPostsForChannel(channelID, 0, 1)
	if appErr != nil {
		return "", appErr
	}

	if len(posts.Posts) == 0 {
		return "", errors.New("Previous post not found")
	}

	var post *model.Post
	for _, p := range posts.Posts {
		post = p
	}

	if len(post.FileIds) == 0 {
		return "", errors.New("No file found on previous post")
	}

	return post.FileIds[0], nil
}
