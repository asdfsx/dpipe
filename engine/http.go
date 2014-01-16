package engine

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io"
	"net"
	"net/http"
)

func (this *EngineConfig) launchHttpServ() {
	this.httpRouter = mux.NewRouter()
	this.httpServer = &http.Server{Addr: this.String("http_addr", "127.0.0.1:9876"), Handler: this.httpRouter}
	this.addHttpHandlers()

	var err error
	this.listener, err = net.Listen("tcp", this.httpServer.Addr)
	if err != nil {
		panic(err)
	}
	go this.httpServer.Serve(this.listener)

	Globals().Printf("Listening on http://%s", this.httpServer.Addr)
}

func (this *EngineConfig) handleHttpQuery(w http.ResponseWriter, req *http.Request,
	params map[string]interface{}) (interface{}, error) {
	vars := mux.Vars(req)
	cmd := vars["cmd"]
	globals := Globals()
	if globals.Verbose {
		globals.Println(req.Method, cmd)
	}

	output := make(map[string]interface{})
	switch cmd {
	case "ping":
		output["status"] = "ok"

	case "stat":
		output["projects"] = this.projects
		output["totalM"] = this.router.totalProcessedMsgN
		output["periodM"] = this.router.periodProcessMsgN
		output["start"] = globals.StartedAt
		output["pid"] = this.pid
		output["hostname"] = this.hostname

	case "plugins":
		output["plugins"] = this.pluginNames()

	case "paths":
		output["all"] = this.httpPaths
	}

	return output, nil
}

func (this *EngineConfig) addHttpHandlers() {
	this.HttpApiHandleFunc("/admin/{cmd}",
		func(w http.ResponseWriter, req *http.Request,
			params map[string]interface{}) (interface{}, error) {
			return this.handleHttpQuery(w, req, params)
		}).Methods("GET")
}

func (this *EngineConfig) HttpApiHandleFunc(path string,
	handlerFunc func(http.ResponseWriter,
		*http.Request, map[string]interface{}) (interface{}, error)) *mux.Route {
	wrappedFunc := func(w http.ResponseWriter, req *http.Request) {
		var ret interface{}
		params, err := this.decodeHttpParams(w, req)
		if err == nil {
			ret, err = handlerFunc(w, req, params)
		}

		if err != nil {
			ret = map[string]interface{}{"error": err.Error()}
		}

		w.Header().Set("Content-Type", "application/json")
		var status int
		if err == nil {
			status = http.StatusOK
		} else {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)

		if ret != nil {
			// pretty write json result
			pretty, _ := json.MarshalIndent(ret, "", "    ")
			w.Write(pretty)
		}
	}

	this.httpPaths = append(this.httpPaths, path)
	return this.httpRouter.HandleFunc(path, wrappedFunc)
}

func (this *EngineConfig) decodeHttpParams(w http.ResponseWriter, req *http.Request) (map[string]interface{},
	error) {
	params := make(map[string]interface{})
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&params)
	if err != nil && err != io.EOF {
		return nil, err
	}

	return params, nil
}

func (this *EngineConfig) stopHttpServ() {
	if this.listener != nil {
		this.listener.Close()
		Globals().Println("HTTP stopped")
	}
}
