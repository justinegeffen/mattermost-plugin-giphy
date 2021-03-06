package main

import (
	"net/http"
	"strings"
	"sync"

	pluginConf "github.com/moussetc/mattermost-plugin-giphy/server/internal/configuration"
	pluginError "github.com/moussetc/mattermost-plugin-giphy/server/internal/error"
	provider "github.com/moussetc/mattermost-plugin-giphy/server/internal/provider"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"

	"github.com/pkg/errors"
)

const (
	contextKeywords = "keywords"
	contextCaption  = "caption"
	contextGifURL   = "gifURL"
	contextCursor   = "cursor"
	contextRootId   = "rootId"
)

// Plugin is a Mattermost plugin that adds a /gif slash command
// to display a GIF based on user keywords.
type Plugin struct {
	plugin.MattermostPlugin

	configurationLock sync.RWMutex
	configuration     *pluginConf.Configuration

	errorGenerator pluginError.PluginError
	gifProvider    provider.GifProvider
	httpHandler    pluginHTTPHandler
	botId          string
}

// OnActivate register the plugin commands
func (p *Plugin) OnActivate() error {
	if err := p.OnConfigurationChange(); err != nil {
		return errors.Wrap(err, "Could not load plugin configuration")
	}
	p.httpHandler = &defaultHTTPHandler{}
	return p.RegisterCommands()
}

// ExecuteCommand dispatch the command based on the trigger word
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	config := p.getConfiguration()

	if strings.HasPrefix(args.Command, "/"+config.CommandTriggerGifWithPreview) {
		keywords, caption, parseErr := parseCommandLine(args.Command, config.CommandTriggerGifWithPreview)
		if parseErr != nil {
			return nil, p.errorGenerator.FromMessage(parseErr.Error())
		}
		return p.executeCommandGifWithPreview(keywords, caption, args)
	}
	if strings.HasPrefix(args.Command, "/"+config.CommandTriggerGif) {
		keywords, caption, parseErr := parseCommandLine(args.Command, config.CommandTriggerGif)
		if parseErr != nil {
			return nil, p.errorGenerator.FromMessage(parseErr.Error())
		}
		return p.executeCommandGif(keywords, caption)
	}

	return nil, p.errorGenerator.FromMessage("Command trigger " + args.Command + "is not supported by this plugin.")
}

// ServeHTTP serve the post actions for the shuffle command
func (p *Plugin) ServeHTTP(c *plugin.Context, w http.ResponseWriter, r *http.Request) {
	p.handleHTTPRequest(w, r)
}
