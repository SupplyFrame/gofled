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