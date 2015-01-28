(function() {
	var tick = 0;

	var Viewer = function(containerId) {
		this.containerId = containerId;
		this.init();
	}

	Viewer.prototype = {
		init: function() {
			this.viewWidth = 640;
			this.viewHeight = 380;
			this.pixelWidth = 54;
			this.pixelHeight = 32;
			this.pixelSize = 16;

			var renderer = new PIXI.autoDetectRenderer(this.viewWidth, this.viewHeight);
			renderer.view.className = "renderer-view";
			this.renderer = renderer;

			var renderContainer = document.getElementById("render");
			renderContainer.appendChild(renderer.view);

			// create stage and background texture
			var stage = new PIXI.Stage(0x000000);
			var bgTex = PIXI.Texture.fromImage("/public/bg.jpg");
			var bgSprite = new PIXI.Sprite(bgTex);
			stage.addChild(bgSprite);

			this.stage = stage;

			this.initPixels();

			requestAnimationFrame(this.animate.bind(this));

			this.startSocket();
		},
		initPixels: function() {
			var pixels = [];

			var pixelSpacing = Math.min(Math.floor(this.viewWidth/this.pixelWidth), Math.floor(this.viewHeight/this.pixelHeight));

			var pixelXOffset = (this.viewWidth - (pixelSpacing*(this.pixelWidth-1))) / 2,
				pixelYOffset = (this.viewHeight - (pixelSpacing*(this.pixelHeight-1))) / 2;

			// create all the pixels
			for (var x=0; x < this.pixelWidth; x++) {
				pixels[x] = [];
				for (var y=0; y < this.pixelHeight; y++) {
					var pixel = new PIXI.Sprite.fromImage("/public/point.png");
					pixels[x][y] = pixel;
					pixel.position.x = pixelXOffset + (pixelSpacing * x);
					pixel.position.y = pixelYOffset + (pixelSpacing * y);
					pixel.blendMode = PIXI.blendModes.ADD;
					pixel.anchor.x = 0.5;
					pixel.anchor.y = 0.5;

					this.stage.addChild(pixel);
				}
			}

			this.pixelXOffset = pixelXOffset;
			this.pixelYOffset = pixelYOffset;
			this.pixelSpacing = pixelSpacing;
			this.pixels = pixels;
		},
		updatePixels: function(data) {
			var pixels = this.pixels;
			// update the pixels with the new tints
			for (var x=0; x < pixels.length; x++) {
				for (var y=0; y < pixels[x].length; y++) {
					var pixel = pixels[x][y];
					var index = (x + (y*pixels[x].length))*3;
					var r = data[index],
						g = data[index+1],
						b = data[index+2];

					pixel.tint = r << 16 | g << 8 | b;
				}
			}
		},
		animate: function() {
		    this.renderer.render(this.stage);
			requestAnimationFrame(this.animate.bind(this));
		},

		startSocket: function() {
			var c = new WebSocket("ws://" + window.location.hostname + ":9000/client");
			c.onerror = function(e) {
				console.log("Socket error : ", e);
			};

			c.binaryType = 'arraybuffer';

			var that = this;
			c.onopen = function() {
				c.onmessage = function(response) {
					//console.log("Data : ", response.data);
					that.updatePixels(new Uint8Array(response.data));
				};
			};
		}
	};



	window.Viewer = Viewer;

	window.view = new Viewer();
})();