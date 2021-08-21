package main

import (
	"github.com/ipfs/go-ipfs/plugin"
	coreiface "github.com/ipfs/interface-go-ipfs-core"

	"github.com/valist-io/go-ipfs-git/transport/http"
)

// Plugins is exported list of plugins that will be loaded
var Plugins = []plugin.Plugin{
	&GitPlugin{},
}

var _ plugin.PluginDaemon = (*GitPlugin)(nil)

type GitPlugin struct{}

func (p *GitPlugin) Name() string {
	return "ipfs-git"
}

func (p *GitPlugin) Version() string {
	return "0.0.1"
}

func (p *GitPlugin) Init(env *plugin.Environment) error {
	// TODO configure server settings
	return nil
}

func (p *GitPlugin) Start(api coreiface.CoreAPI) error {
	go http.ListenAndServe(api, ":8081")
	return nil
}

func (p *GitPlugin) Close() error {
	// TODO shutdown http server gracefully
	return nil
}
