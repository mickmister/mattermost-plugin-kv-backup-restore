package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/pkg/errors"
)

type KVWithString struct {
	PluginId string
	Key      string
	Value    string
	ExpireAt int64
}

func executeBackupSQL(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	db, err := p.client.Store.GetMasterDB()
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "failed to get a database connection").Error())
	}

	q := fmt.Sprintf(`
		SELECT PluginId, PKey, PValue, ExpireAt FROM PluginKeyValueStore WHERE PluginId = '%s'
	`, args[0])

	rows, err := db.Query(q)
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "error querying database").Error())
	}

	allValues := map[string]*KVWithString{}

	defer rows.Close()
	for rows.Next() {
		var id string
		var key string
		var value []byte
		var expireAt int64

		err := rows.Scan(&id, &key, &value, &expireAt)
		if err != nil {
			return p.responsef(cmdArgs, errors.Wrap(err, "error getting db row").Error())
		}
		s := base64.StdEncoding.EncodeToString(value)
		kv := &KVWithString{
			PluginId: id,
			Key:      key,
			Value:    s,
			ExpireAt: expireAt,
		}
		allValues[key] = kv
	}

	b, err := json.Marshal(allValues)
	if err != nil {
		return p.responsef(cmdArgs, errors.Wrap(err, "error marshaling response").Error())
	}

	post := &model.Post{
		UserId:    cmdArgs.UserId,
		ChannelId: cmdArgs.ChannelId,
	}

	ts := time.Now().Format(time.Kitchen)
	fname := fmt.Sprintf("%s-backup-%s.json", manifest.Id, ts)

	fileInfo, appErr := p.API.UploadFile(b, cmdArgs.ChannelId, fname)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error uploading result err=%v", appErr)
	}

	post.FileIds = append(post.FileIds, fileInfo.Id)
	res := fmt.Sprintf("Backed up %d values", len(allValues))

	post.Message = res
	p.API.CreatePost(post)
	return &model.CommandResponse{}
}

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

		s := string(value)
		if isGeneratedKeyValue(key) || true {
			s = `"` + base64.StdEncoding.EncodeToString(value) + `"`
		} else if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			var buf bytes.Buffer
			err := json.Indent(&buf, value, "  ", "  ")
			if err == nil {
				s = string(buf.Bytes())
			}
		} else {
			b, err := json.Marshal(s)
			if err != nil {
				return p.responsef(cmdArgs, "Error marshaling value for key `%s`. err=%v", key, err)
			}

			s = string(b)
		}

		comma := ",\n"
		if i == len(keys)-1 {
			comma = ""
		}
		allValues += fmt.Sprintf("  \"%s\": %s%s", key, s, comma)
	}

	asJSON := fmt.Sprintf("{\n%s\n}", allValues)

	if len(args) == 0 || args[0] != "file" {
		res := fmt.Sprintf("```json\n%s\n```", asJSON)
		return p.responsef(cmdArgs, res)
	}

	post := &model.Post{
		UserId:    cmdArgs.UserId,
		ChannelId: cmdArgs.ChannelId,
	}

	ts := time.Now().Format(time.Kitchen)
	fname := fmt.Sprintf("%s-backup-%s.json", manifest.Id, ts)

	fileInfo, appErr := p.API.UploadFile([]byte(asJSON), cmdArgs.ChannelId, fname)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error uploading result err=%v", appErr)
	}

	post.FileIds = append(post.FileIds, fileInfo.Id)
	res := fmt.Sprintf("Backed up %d values", len(keys))

	post.Message = res
	p.API.CreatePost(post)
	return &model.CommandResponse{}
}
