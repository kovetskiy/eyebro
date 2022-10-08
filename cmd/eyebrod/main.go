package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/docopt/docopt-go"
	"github.com/gorilla/websocket"
	"github.com/kovetskiy/eyebro/internal/config"
	"github.com/reconquest/pkg/log"
)

var (
	version = "[manual build]"
	usage   = "eyebrod " + version + os.ExpandEnv(`

Usage:
  eyebrod [options]
  eyebrod -h | --help
  eyebrod --version

Options:
  -h --help           Show this screen.
  -c --config <path>  Use specified config file [default: $HOME/.config/eyebrod.conf].
  --version           Show version.
`)
)

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	config, err := config.Load(args["--config"].(string))
	if err != nil {
		log.Fatalf(err, "unable to load config")
	}

	bus := NewBus()
	http.Handle("/ws", &WebSocket{bus: bus})
	http.Handle("/rpc", &RPC{bus: bus})

	log.Infof(nil, "listening on %s", config.Listen)
	err = http.ListenAndServe(config.Listen, nil)
	if err != nil {
		log.Fatal(err)
	}
}

type WebSocket struct {
	bus *Bus
}

func (socket *WebSocket) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(*http.Request) bool { return true },
	}

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf(err, "unable to upgrade connection")
		return
	}
	defer c.Close()

	requests, _ := socket.bus.Subscribe("request")
	defer socket.bus.Close("request")

	log.Infof(nil, "browser connected")

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Errorf(err, "unable to read message")
				break
			}

			socket.bus.Publish("response", string(message))
		}
	}()

	for {
		select {
		case request, ok := <-requests:
			if !ok {
				return
			}

			message := request.(string)

			err := c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				return
			}
		}
	}
}

type RPC struct {
	bus *Bus
}

func (rpc *RPC) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	command := r.URL.Query().Get("command")
	if command == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var args interface{}
	rawArgs := r.URL.Query().Get("args")
	if rawArgs != "" {
		err := json.Unmarshal([]byte(rawArgs), &args)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	if rpc.bus.Len("request") == 0 {
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	responses, _ := rpc.bus.Subscribe("response")
	defer rpc.bus.Close("response")

	rpc.bus.Publish("request", encodeJSON(map[string]interface{}{
		"command": command,
		"args":    args,
	}))

	response := <-responses
	w.Write([]byte(response.(string)))
}

func encodeJSON(obj interface{}) string {
	data, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}

	return string(data)
}
