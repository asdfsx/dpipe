package engine

import (
	"fmt"
	conf "github.com/funkygao/jsconf"
	"github.com/funkygao/pretty"
	"github.com/gorilla/mux"
	"net"
	"net/http"
	"os"
	"time"
)

type EngineConfig struct {
	*conf.Conf

	listener   net.Listener
	httpServer *http.Server
	httpRouter *mux.Router

	projects map[string]*ConfProject

	InputRunners  map[string]InputRunner
	inputWrappers map[string]*PluginWrapper

	FilterRunners  map[string]FilterRunner
	filterWrappers map[string]*PluginWrapper

	OutputRunners  map[string]OutputRunner
	outputWrappers map[string]*PluginWrapper

	router *messageRouter

	// PipelinePack supply for Input plugins.
	inputRecycleChan chan *PipelinePack

	// PipelinePack supply for Filter plugins
	injectRecycleChan chan *PipelinePack

	hostname string
	pid      int
}

func NewEngineConfig(globals *GlobalConfigStruct) (this *EngineConfig) {
	this = new(EngineConfig)

	if globals == nil {
		globals = DefaultGlobals()
	}
	Globals = func() *GlobalConfigStruct {
		return globals
	}

	this.InputRunners = make(map[string]InputRunner)
	this.inputWrappers = make(map[string]*PluginWrapper)
	this.FilterRunners = make(map[string]FilterRunner)
	this.filterWrappers = make(map[string]*PluginWrapper)
	this.OutputRunners = make(map[string]OutputRunner)
	this.outputWrappers = make(map[string]*PluginWrapper)

	this.inputRecycleChan = make(chan *PipelinePack, globals.PoolSize)
	this.injectRecycleChan = make(chan *PipelinePack, globals.PoolSize)

	this.projects = make(map[string]*ConfProject)

	this.router = NewMessageRouter()

	this.hostname, _ = os.Hostname()
	this.pid = os.Getpid()

	return this
}

func (this *EngineConfig) EngineConfig() *EngineConfig {
	return this
}

func (this *EngineConfig) Project(name string) *ConfProject {
	p, present := this.projects[name]
	if !present {
		panic("invalid project: " + name)
	}

	return p
}

// For Filter to generate new pack
func (this *EngineConfig) PipelinePack(msgLoopCount int) *PipelinePack {
	if msgLoopCount++; msgLoopCount > Globals().MaxMsgLoops {
		return nil
	}

	pack := <-this.injectRecycleChan
	pack.Reset()
	pack.MsgLoopCount = msgLoopCount

	return pack
}

func (this *EngineConfig) LoadConfigFile(fn string) {
	cf, err := conf.Load(fn)
	if err != nil {
		panic(err)
	}

	this.Conf = cf

	globals := Globals()
	if globals.Debug {
		globals.Printf("config: %#v\n", *cf)
	}

	// 'projects' section
	for i := 0; i < len(this.List("projects", nil)); i++ {
		section, err := this.Section(fmt.Sprintf("projects[%d]", i))
		if err != nil {
			panic(err)
		}

		project := &ConfProject{}
		project.FromConfig(section)
		if _, present := this.projects[project.Name]; present {
			panic("dup project: " + project.Name)
		}
		this.projects[project.Name] = project
	}

	// 'plugins' section
	for i := 0; i < len(this.List("plugins", nil)); i++ {
		section, err := this.Section(fmt.Sprintf("plugins[%d]", i))
		if err != nil {
			panic(err)
		}

		this.loadPluginSection(section)
	}
}

func (this *EngineConfig) loadPluginSection(section *conf.Conf) {
	pluginCommons := new(pluginCommons)
	pluginCommons.load(section)
	if pluginCommons.disabled {
		globals := Globals()
		if globals.Verbose {
			globals.Printf("[%s] disabled\n", pluginCommons.name)
		}

		return
	}

	wrapper := new(PluginWrapper)
	var ok bool
	if wrapper.pluginCreator, ok = availablePlugins[pluginCommons.class]; !ok {
		pretty.Printf("allPlugins: %# v\n", availablePlugins)
		panic("unknown plugin type: " + pluginCommons.class)
	}
	wrapper.configCreator = func() *conf.Conf { return section }
	wrapper.name = pluginCommons.name

	plugin := wrapper.pluginCreator()
	plugin.Init(section)

	pluginCats := pluginTypeRegex.FindStringSubmatch(pluginCommons.class)
	if len(pluginCats) < 2 {
		panic("invalid plugin type: " + pluginCommons.class)
	}

	pluginCategory := pluginCats[1]
	if pluginCategory == "Input" {
		this.InputRunners[wrapper.name] = NewInputRunner(wrapper.name, plugin.(Input),
			pluginCommons)
		this.inputWrappers[wrapper.name] = wrapper
		if pluginCommons.ticker > 0 {
			this.InputRunners[wrapper.name].setTickLength(time.Duration(pluginCommons.ticker) * time.Second)
		}

		return
	}

	runner := NewFORunner(wrapper.name, plugin, pluginCommons)
	matcher := NewMatchRunner(section.StringList("match", []string{}), runner)
	runner.matcher = matcher

	switch pluginCategory {
	case "Filter":
		this.router.filterMatchers = append(this.router.filterMatchers, matcher)
		this.FilterRunners[runner.name] = runner
		this.filterWrappers[runner.name] = wrapper

	case "Output":
		this.router.outputMatchers = append(this.router.outputMatchers, matcher)
		this.OutputRunners[runner.name] = runner
		this.outputWrappers[runner.name] = wrapper
	}
}

// common config directives for all plugins
type pluginCommons struct {
	name     string `json:"name"`
	class    string `json:"class"`
	poolSize int    `json:"pool_size"`
	ticker   int    `json:"ticker_interval"`
	disabled bool   `json:"disabled"`
	comment  string `json:"comment"`
}

func (this *pluginCommons) load(section *conf.Conf) {
	this.name = section.String("name", "")
	if this.name == "" {
		pretty.Printf("%# v\n", *section)
		panic(fmt.Sprintf("invalid plugin config: %v", *section))
	}

	this.class = section.String("class", "")
	if this.class == "" {
		this.class = this.name
	}
	this.comment = section.String("comment", "")
	this.poolSize = section.Int("pool_size", Globals().PoolSize)
	this.ticker = section.Int("ticker_interval", Globals().TickerLength)
	this.disabled = section.Bool("disabled", false)
}
