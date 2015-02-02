window.views[0] = new Viewer("render");
	
var c = new WebSocket("ws://" + window.location.hostname + ":9000/client");
c.onerror = function(e) {
	console.log("Socket error : ", e);
};

c.binaryType = 'arraybuffer';

var activeSource = null;
c.onopen = function() {
	c.onmessage = function(response) {
		if (typeof response.data === "string") {
			// this is a normal json message
			var msg = JSON.parse(response.data);
			if (msg.Type == "data") {
				if (msg.ID == "active") {
					activeSource = 0;
				} else {
					activeSource = msg.ID;
				}
			}
			if (msg.Type == "add-source") {
				// add a new source to our sources div and create a viewer on it
				var div = document.createElement("div");
				div.className = "source";
				div.id = "source-" + msg.ID;
				document.getElementById("sources").appendChild(div);
				var view = new Viewer("source-" + msg.ID);
				window.views[msg.ID] = view;
				console.log("Source added : " + JSON.stringify(msg));
			} else if (msg.Type == "del-source") {
				// remove the viewer and div for this source
				var el = document.getElementById("source-" + msg.ID);
				if (!el ) {
					return // couldn't find the source to remove
				}
				el.parentNode.removeChild(el);
				delete(window.views[msg.ID]);
				window.views[msg.ID] = null;
			}
		} else {
			if (activeSource!==null) {
				// use the first byte to figure out which source this is from
				var rawData = new Uint8Array(response.data);
				var view = window.views[activeSource];
				if (view) {
					view.updatePixels(rawData);
				}
			}
		}
	};
};