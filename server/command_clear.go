package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

func executeClear(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	pluginID := args[0]
	err := p.deleteAllKeys(pluginID)

	if err != nil {
		return p.responsef(cmdArgs, "Error deleting keys. err=%v", err)
	}

	p.responsef(cmdArgs, "Successfully cleared the kv store")
	return executeClearConfig(p, c, cmdArgs, args...)
}

func executeClearKV(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	pluginID := args[0]
	err := p.deleteAllKeys(pluginID)

	if err != nil {
		return p.responsef(cmdArgs, "Error deleting keys. err=%v", err)
	}

	return p.responsef(cmdArgs, "Successfully cleared the kv store")
}

func executeClearConfig(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	config := p.API.GetConfig()
	if config == nil {
		return p.responsef(cmdArgs, "Error getting config")
	}

	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	pluginID := args[0]
	if config.PluginSettings.Plugins[pluginID] == nil {
		return p.responsef(cmdArgs, "Plugin not found")
	}

	config.PluginSettings.Plugins[pluginID] = map[string]interface{}{}

	appErr := p.API.SaveConfig(config)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error setting config. err=%v", appErr)
	}

	return p.responsef(cmdArgs, "Successfully cleared the config")
}

func executeGetConfig(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	config := p.API.GetConfig()
	if config == nil {
		return p.responsef(cmdArgs, "Error getting config")
	}

	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	pluginID := args[0]
	if config.PluginSettings.Plugins[pluginID] == nil {
		return p.responsef(cmdArgs, "Plugin not found")
	}

	b, err := json.MarshalIndent(config.PluginSettings.Plugins[pluginID], "", "  ")
	if err != nil {
		return p.responsef(cmdArgs, "Failed to marshal config")
	}

	out := fmt.Sprintf("```json\n%s\n```", string(b))

	return p.responsef(cmdArgs, out)
}

func executeResetPlugin(p *Plugin, c *plugin.Context, cmdArgs *model.CommandArgs, args ...string) *model.CommandResponse {
	if len(args) == 0 {
		return p.responsef(cmdArgs, "Please provide a plugin ID")
	}

	pluginID := args[0]
	appErr := p.API.DisablePlugin(pluginID)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error disabling plugin. err=%v", appErr)
	}

	p.responsef(cmdArgs, "Successfully disabled plugin")

	time.Sleep(time.Second * 2)
	appErr = p.API.EnablePlugin(pluginID)
	if appErr != nil {
		return p.responsef(cmdArgs, "Error enabling plugin. err=%v", appErr)
	}

	return p.responsef(cmdArgs, "Successfully enabled plugin")
}
