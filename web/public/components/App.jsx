import React from 'react';
import BufferedRecorder from './BufferedRecorder.jsx';

require('../css/App.scss');

const wsAddr = "ws://127.0.0.1:8080/ws";
const bufferSize = 4096;
const actions = {
	INIT: 0,
	STARTING: 1,
	STOPPING: 2,
};
const packets = {
	SPECTRAL: 0,
	RESULT: 1,
	PEAK: 2,
};

class App extends React.Component {
	constructor(props) {
		super(props);

		this.sendBuffer = this.sendBuffer.bind(this);
		this.onRecordClick = this.onRecordClick.bind(this);

		this.state = {
			recording: false,
			canvasX: 0,
		};

		// TODO(sno6): Should this be state?
		this.binCount = 0;
	}

	initSocket(callback) {
		this.socket = new WebSocket(wsAddr);

		this.socket.onopen = () => {
			let msg = JSON.stringify({ action: actions.INIT, bufferSize: bufferSize });
			this.socket.send(msg);

			// Initialise the canvas.
			this.spec = document.getElementById("spectrogram");
			this.ctx = this.spec.getContext("2d");

			// Effects the detail of the canvas.
			this.spec.height = 1024 / 2;
			this.spec.width = 900;

			callback();
		};

		this.socket.onmessage = (m) => {
			let data = JSON.parse(m.data);
			switch (data.type) {
				case packets.SPECTRAL:
					this.binCount += data.bins.length;
					this.drawSpectral(data);
					break;
				case packets.PEAK:
					if (typeof data.peaks != 'undefined') {
						this.drawPeaks(data);
					}
			}
		};

		this.socket.onerror = (e) => {
			console.log(e);
		};

		this.socket.onclose = () => {
			console.log("Closing Websocket");
		};
	}

	sendBuffer(buffer) {
		this.socket.send(buffer);
	}

	onRecordClick(recording) {
		this.setState({ recording: recording, canvasX: 0 });

		if (recording) {
			if (this.socket != null) {
				return;
			}
			this.initSocket(() => {
				let msg = JSON.stringify({ action: actions.STARTING });
				this.socket.send(msg)
			});
		} else {
			let msg = JSON.stringify({ action: actions.STOPPING });
			this.socket.send(msg);
			this.socket.close();
			this.socket = null;
		}
	}

	drawSpectral(data) {
		// For each of the four frequency bins of ~1024 samples.
		for (var i = 0; i < data.bins.length; i++) {
			// Get the current bin.
			let bin = data.bins[i];

			// Creates a data array of width*height*4 or 1*1024*4 colour values.
			let imgData = this.ctx.createImageData(1, 1024 / 2);

			for (var j = 0; j < imgData.data.length; j += 4) {
				let mag = bin[j / 4];
				let color = this.getColor(mag);
				imgData.data[j + 0] = color.r;
				imgData.data[j + 1] = color.g;
				imgData.data[j + 2] = color.b;
				imgData.data[j + 3] = 255;
			}

			this.ctx.putImageData(imgData, this.state.canvasX + i, 0);
		}
		this.setState(prevState => ({ canvasX: prevState.canvasX + data.bins.length }));

		if (this.state.canvasX > this.spec.width) {
			this.setState({ canvasX: 0 });
		}
	}

	drawPeaks(data) {
		for (var i = 0; i < data.peaks.length; i++) {
			let peak = data.peaks[i];

			// Find the offset from the peak and the canvas so we know
			// where to place the peak visual.
			let delta = this.binCount - peak.time;
			let x = this.state.canvasX - delta;

			this.drawCircle(x, peak.freq, 2);
		}
	}

	drawCross(x, y, length) {
		this.ctx.beginPath();
		this.ctx.strokeStyle = 'black';

		this.ctx.moveTo(x - length, y - length);
		this.ctx.lineTo(x + length, y + length);
		this.ctx.stroke();

		this.ctx.moveTo(x + length, y - length);
		this.ctx.lineTo(x - length, y + length);
		this.ctx.stroke();
	}

	drawCircle(x, y, radius) {
		this.ctx.beginPath();
		this.ctx.arc(x, y, radius, 0, 2 * Math.PI, false);
		this.ctx.lineWidth = 1;
		this.ctx.fillStyle = '#000000';
		this.ctx.fill();
	}

	getColor(mag) {
		let max = 100;
		let power = -1 * this.magToPower(mag, 1024 / 2);
		let color = this.interpolateColor(power, max);
		return color;
	}

	interpolateColor(power, max) {
		if (power > max) {
			power = max;
		} else {
			power = power / max;
		}

		let spectrum = [
			{ color: { r: 0, g: 0, b: 0 }, value: 1.0 },
			{ color: { r: 1, g: 3, b: 63 }, value: 0.9 },
			{ color: { r: 97, g: 14, b: 125 }, value: 0.8 },
			{ color: { r: 145, g: 14, b: 119 }, value: 0.7 },
			{ color: { r: 187, g: 13, b: 91 }, value: 0.6 },
			{ color: { r: 168, g: 14, b: 107 }, value: 0.5 },
			{ color: { r: 200, g: 12, b: 77 }, value: 0.4 },
			{ color: { r: 233, g: 11, b: 30 }, value: 0.3 },
			{ color: { r: 253, g: 149, b: 38 }, value: 0.2 },
			{ color: { r: 250, g: 204, b: 152 }, value: 0.1 },
			{ color: { r: 255, g: 231, b: 202 }, value: 0.0 },
		];

		// Find the color between two color values.
		for (let c = 0; c < spectrum.length; c++) {
			if (power == spectrum[c].value) {
				return spectrum[c].color;
			}

			// In between two values.
			if (power < spectrum[c].value && power > spectrum[c + 1].value) {
				let left = spectrum[c].color;
				let right = spectrum[c + 1].color;
				let weight = (power - spectrum[c + 1].value) * 10;
				let col = chroma.mix([right.r, right.g, right.b], [left.r, left.g, left.b], weight, 'rgb').rgb();
				return { r: col[0], g: col[1], b: col[2] };
			}
		}

		return { r: 0, g: 0, b: 0 };
	}

	magToPower(mag, windowSize) {
		windowSize = windowSize / 2
		return 10 * Math.log10((2 * mag) / windowSize);
	}

	openMenu() {
		$('.menu-transform').toggleClass('menu-transform-active');
	}

	render() {
		let spectrogram = null;
		if (this.state.recording) {
			spectrogram = <canvas id="spectrogram" className="spectrogram"></canvas>;
		}

		return (
			<div className="container">
				<div className="app">
					<div className="header">
						<p id="filler"></p>
						<p>Columbo</p>
						<i id="settings" class="fa fa-sliders" aria-hidden="true" onClick={() => this.openMenu()}></i>

					</div>
					<div className="content-container">
						{spectrogram}
						<BufferedRecorder bufferSize={bufferSize} onClick={this.onRecordClick} onBufferCallback={this.sendBuffer} />
					</div>
				</div>

				<div className="menu menu-transform">
					<p class="menu-header">VISUAL</p>
					<div class="menu-item">
						<p>Show peaks</p>
						<label class="switch">
							<input type="checkbox" />
							<span class="slider round"></span>
						</label>
					</div>

					<p class="menu-header">BACKEND</p>
					<div class="menu-item">
						<p>Add more options</p>
					</div>

				</div >
			</div>
		);
	}
}

export default App;