const addr = 'ws://localhost:3451/ws'

function connect() {
    websocket = new WebSocket(addr);

    websocket.onopen = function() {};

    websocket.onmessage = async function (event) {
        const packet = JSON.parse(event.data);
        const command = packet["command"];
        const args = packet["args"];
        if (command === "get_hostname") {
            chrome.tabs.getSelected(null, function(tab) {
                let hostname = (new URL(tab.url)).hostname;
                hostname = hostname.replace(/^(www\.)/, "")
                websocket.send(hostname);
            });
        } else if (command === "list_tabs") {
            await chrome.tabs.query({}, function(tabs) {
                let urls = tabs.map(function(tab) {
                    return {"id": tab.id, "url": tab.url, "title": tab.title};
                });

                websocket.send(JSON.stringify(urls));
            })
        } else if (command === "switch_tab") {
            const id = args;
            chrome.tabs.update(id, {"active": true});
            websocket.send("");
        } else {
            websocket.send("unknown command: "+command);
        }
    };

    websocket.onclose = function() {
        websocket = undefined;
        setTimeout(connect, 1000);
    };
}

connect();
