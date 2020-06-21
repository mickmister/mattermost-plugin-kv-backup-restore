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
	parentCmd := model.NewAutocompleteData("kvadmin", "(pluginid) [list|show|put|backup|restore|help]", "Manage and Backup/Restore data in your kv store.")
	cmd := model.NewAutocompleteData("(pluginid)", "[list|show|put|backup|restore]", "Manage and Backup/Restore data in your kv store.")

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
	parentCmd.AddCommand(cmd)

	return &model.Command{
		Trigger:          "kvadmin",
		Description:      "Manage and Backup/Restore data in your kv store.",
		DisplayName:      "KV Backup/Restore",
		AutoComplete:     true,
		AutocompleteData: parentCmd,
		AutoCompleteDesc: parentCmd.HelpText,
		AutoCompleteHint: parentCmd.Hint,
	}
}

func executeDefault(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	return p.responsef(cmdArgs, "Provide a command!")
}

func (p *Plugin) ExecuteCommand(c *plugin.Context, cmdArgs *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	args := strings.Fields(cmdArgs.Command)
	if len(args) == 0 || args[0] != "/kvadmin" {
		return p.responsef(cmdArgs, "expected kvadmin command"), nil
	}

	if len(args) == 1 || args[1] != manifest.Id {
		return p.responsef(cmdArgs, "Expected plugin id `%s`", manifest.Id), nil
	}

	if len(args) == 2 {
		return p.responsef(cmdArgs, p.getHelpText()), nil
	}

	return cmdHandler.Handle(p, c, cmdArgs, args[2:]...), nil
}

func (p *Plugin) responsef(commandArgs *model.CommandArgs, format string, args ...interface{}) *model.CommandResponse {
	p.postCommandResponse(commandArgs, fmt.Sprintf(format, args...))
	return &model.CommandResponse{}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		// UserId:    p.getUserID(),
		UserId:    args.UserId,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	_ = p.API.SendEphemeralPost(args.UserId, post)
}

func webappSlashCommandWillBePostedHook() {
	// add e2e encryption token at the front of the command
}

func (p *Plugin) getHelpText() string {
	return "Help Text"
}
