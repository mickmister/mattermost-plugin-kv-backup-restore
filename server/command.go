package main

import (
	"fmt"
	"strings"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

type CommandHandlerFunc func(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse

type CommandHandler struct {
	handlers       map[string]CommandHandlerFunc
	defaultHandler CommandHandlerFunc
}

var cmdHandler = CommandHandler{
	handlers: map[string]CommandHandlerFunc{
		"list":    executeList,
		"show":    executeShow,
		"update":  executeUpdate,
		"backup":  executeBackup,
		"restore": executeRestore,
		"delete":  executeDelete,
		"clear":   executeClear,
	},
	defaultHandler: executeDefault,
}

func (ch CommandHandler) Handle(p *Plugin, c *plugin.Context, header *model.CommandArgs, args ...string) *model.CommandResponse {
	for n := len(args); n > 0; n-- {
		h := ch.handlers[strings.Join(args[:n], "/")]
		if h != nil {
			return h(p, c, header, args[n:]...)
		}
	}
	return ch.defaultHandler(p, c, header, args...)
}

func newCommand(pluginID string) *model.Command {
	trigger := "kvadmin-" + pluginID

	cmd := model.NewAutocompleteData(trigger, "[list|show|put|backup|restore]", "Manage and Backup/Restore data in your kv store.")

	listCmd := model.NewAutocompleteData("list", "[keys|table]", "List all keys in the kvstore.")
	showCmd := model.NewAutocompleteData("show", "[key]", "Show the value of one kv entry.")
	updateCommand := model.NewAutocompleteData("update", "[key] [value]", "Update one kv entry's value.")
	backupCmd := model.NewAutocompleteData("backup", "", "Receive a json blob of the whole kvstore for this plugin.")
	restoreCmd := model.NewAutocompleteData("restore", "[json blob]", "Set all key values in the kvstore.")

	cmd.AddCommand(listCmd)
	cmd.AddCommand(showCmd)
	cmd.AddCommand(updateCommand)
	cmd.AddCommand(backupCmd)
	cmd.AddCommand(restoreCmd)

	return &model.Command{
		Trigger:          trigger,
		Description:      "Manage and Backup/Restore data in your kv store.",
		DisplayName:      "KV Backup/Restore",
		AutoComplete:     true,
		AutocompleteData: cmd,
		AutoCompleteDesc: cmd.HelpText,
		AutoCompleteHint: cmd.Hint,
	}
}

func executeDefault(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	return p.responsef(cmdArgs, "Provide a command!")
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, cmdArgs *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	args := strings.Fields(cmdArgs.Command)
	if len(args) == 0 || args[0] != "/kvadmin-"+manifest.Id {
		return p.responsef(cmdArgs, "expected kvadmin command"), nil
	}

	if len(args) == 1 {
		return p.responsef(cmdArgs, p.getHelpText()), nil
	}

	return cmdHandler.Handle(p, c, cmdArgs, args[1:]...), nil
}

func (p *Plugin) responsef(commandArgs *model.CommandArgs, format string, args ...interface{}) *model.CommandResponse {
	p.postCommandResponse(commandArgs, fmt.Sprintf(format, args...))
	return &model.CommandResponse{}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func (p *Plugin) getHelpText() string {
	commands := []string{
		"- `list` - Returns a JSON array of all keys in your plugin's kv store",
		"- `show` `(key)` - Returns value associated with the key",
		"- `update` `(key)` `((data) | file (fileID))` - Updates or creates an entry for key. Provide raw data as an arg, or the word `file` along with an uploded file's `fileID`",
		"	- `update` `mykey` `{\"some\": \"value\"}` - Uses raw data from command args",
		"	- `update` `mykey` `file` `17a889qjmjg8zpf7qtys8oy5tw` - Uses previously uploaded file id",
		"- `backup (file?)` -  Returns a JSON object for all key value entries",
		"- `restore` `((data) | file (fileID))` - Similar to update, except for the whole store. First, it clears the store, then adds all contents provided.",
		"	- `restore` `{\"mykey\": {\"some\": \"value\"}, \"other_key\": \"hello\"}` - Uses raw data from command args",
		"	- `restore` `file` `17a889qjmjg8zpf7qtys8oy5tw` - Uses previously uploaded file id",
		"- `delete` `mykey` - Deletes a key from the kv store",
		"- `clear` - Clears all values in the kv store",
	}
	return strings.Join(commands, "\n")
}
