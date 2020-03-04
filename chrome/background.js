const addr = 'ws://localhost:3451/ws'

function connect() {
    websocket = new WebSocket(addr);

    websocket.onopen = function() {};

    websocket.onmessage = function (event) {
        const command = event.data;
        if (command == "get_hostname") {
            chrome.tabs.getSelected(null, function(tab) {
                let hostname = (new URL(tab.url)).hostname;
                hostname = hostname.replace(/^(www\.)/, "")
                websocket.send(hostname);
            });
        } else {
            websocket.send("unknown command");
        }
    };

    websocket.onclose = function() {
        websocket = undefined;
        connect();
    };
}

connect();
