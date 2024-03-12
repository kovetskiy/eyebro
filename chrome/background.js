const addr = "ws://localhost:3451/ws";

function connect() {
  console.log("[background] connecting to " + addr);
  const websocket = new WebSocket(addr);
  chrome.runtime.websocket = websocket;

  websocket.onopen = function () {
    console.log("[background] websocket connected");
  };

  websocket.onmessage = async function (event) {
    const packet = JSON.parse(event.data);
    const command = packet["command"];
    const args = packet["args"];

    console.log("[background] received command: " + command);

    if (command === "get_hostname") {
      const [tab] = await chrome.tabs.query({
        active: true,
        currentWindow: true,
      });
      let hostname = new URL(tab.url).hostname;
      hostname = hostname.replace(/^(www\.)/, "");
      console.log("[background] send: " + hostname);
      websocket.send(hostname);
    } else if (command === "list_tabs") {
      const tabs = await chrome.tabs.query({});
      let urls = tabs.map((tab) => {
        return { id: tab.id, url: tab.url, title: tab.title };
      });

      console.log("[background] send: " + JSON.stringify(urls));
      websocket.send(JSON.stringify(urls));
    } else if (command === "switch_tab") {
      const id = args;
      await chrome.tabs.update(id, { active: true });
      console.log('[background] send: ""');
      websocket.send("");
    } else {
      console.log("[background] send: unknown command: " + command);
      websocket.send("unknown command: " + command);
    }
  };

  websocket.onclose = function () {
    console.log("[background] websocket closed");
    if (chrome.runtime.websocket) {
      chrome.runtime.websocket = null;
      setTimeout(connect, 1000);
    }
  };
}

chrome.runtime.onStartup.addListener(connect);
connect();
