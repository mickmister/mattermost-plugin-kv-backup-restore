package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	pluginapi "github.com/mattermost/mattermost-plugin-api"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

var generatedKeyValues = []string{
	"token_secret",
	"rsa_key",
}

func isGeneratedKeyValue(key string) bool {
	for _, name := range generatedKeyValues {
		if name == key {
			return true
		}
	}

	return false
}

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin

	// configurationLock synchronizes access to the configuration.
	configurationLock sync.RWMutex

	// configuration is the active plugin configuration. Consult getConfiguration and
	// setConfiguration for usage.
	configuration *configuration

	client *pluginapi.Client
}

// ServeHTTP demonstrates a plugin that handles HTTP requests by greeting the world.
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, world!")
}

func (p *Plugin) OnActivate() error {
	manifest = model.ManifestFromJson(strings.NewReader(manifestStr))

	cmd := newCommand(manifest.Id)
	p.API.RegisterCommand(cmd)

	p.client = pluginapi.NewClient(p.API)

	return nil
}
