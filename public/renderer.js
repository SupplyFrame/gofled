(function() {
	var tick = 0;

	var Viewer = function(containerId) {
		this.containerId = containerId;
		this.container = document.getElementById(containerId);
		this.init();
	}

	Viewer.prototype = {
		init: function() {
			this.viewWidth = this.container.clientWidth;
			this.viewHeight = this.container.clientHeight;
			this.pixelWidth = 52;
			this.pixelHeight = 34;

			var renderer = new PIXI.autoDetectRenderer(this.viewWidth, this.viewHeight);
			renderer.view.className = "renderer-view";
			this.renderer = renderer;

			this.container.appendChild(renderer.view);

			var stage = new PIXI.Stage(0x000000);

			this.stage = stage;

			this.pixelContainer = new PIXI.DisplayObjectContainer();
			stage.addChild(this.pixelContainer);

			this.initPixels();

			requestAnimationFrame(this.animate.bind(this));
		},

		
		resize: function() {
			this.viewWidth = this.container.clientWidth;
			this.viewHeight = this.container.clientHeight;

			this.renderer.resize(this.viewWidth, this.viewHeight);

			var renderWidth = 640, renderHeight = 380;

			var pixelSpacing = Math.min(Math.floor(renderWidth/(this.pixelWidth-1)), Math.floor(renderHeight/(this.pixelHeight-1)));

			// ideal rendering is at 640x380
			// figure out scale ratio for this new size
			var scaleX = 1 / (renderWidth / this.viewWidth),
				scaleY = 1 / (renderHeight / this.viewHeight);

			var minScale = Math.min(scaleX,scaleY);
			this.pixelContainer.scale = new PIXI.Point(minScale,minScale);

			var pixelXOffset = (renderWidth - (pixelSpacing*(this.pixelWidth-1))) / 2,
				pixelYOffset = (renderHeight - (pixelSpacing*(this.pixelHeight-1))) / 2;

			// create all the pixels
			for (var y=0; y < this.pixelHeight; y++) {
				for (var x=0; x < this.pixelWidth; x++) {
					pixel = this.pixels[y][x];
					pixel.position.x = pixelXOffset + (pixelSpacing * x);
					pixel.position.y = renderHeight - pixelYOffset - (pixelSpacing * y);
				}
			}

			this.pixelXOffset = pixelXOffset;
			this.pixelYOffset = pixelYOffset;
			this.pixelSpacing = pixelSpacing;
		},
		initPixels: function() {
			var pixels = [];

			// create all the pixels
			for (var y=0; y < this.pixelHeight; y++) {
				pixels[y] = [];
				for (var x=0; x < this.pixelWidth; x++) {
					var pixel = new PIXI.Sprite.fromImage("/public/point-small.png");
					pixels[y][x] = pixel;
					pixel.blendMode = PIXI.blendModes.ADD;
					pixel.alpha = 1;
					pixel.anchor.x = 0.5;
					pixel.anchor.y = 0.5;

					this.pixelContainer.addChild(pixel);
				}
			}
			this.pixels = pixels;
			this.resize();
		},
		updatePixels: function(data) {
			var pixels = this.pixels;
			// update the pixels with the new tints
			for (var y=0; y < this.pixelHeight; y++) {
				for (var x=0; x < this.pixelWidth; x++) {
					var pixel = pixels[y][x];
					var index = (x + (y*this.pixelWidth))*3;
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
	};

	window.Viewer = Viewer;
	window.views = {};
})();

