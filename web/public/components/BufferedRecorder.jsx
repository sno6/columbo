import React from 'react';
import PropTypes from 'prop-types';

require('../css/BufferedRecorder.scss');

navigator.getUserMedia = (navigator.getUserMedia ||
    navigator.webkitGetUserMedia ||
    navigator.mozGetUserMedia ||
    navigator.msGetUserMedia);

class BufferedRecorder extends React.Component {
    constructor(props) {
        super(props);

        this.state = {
            recording: false,
        }
    }

    start() {
        navigator.getUserMedia({ audio: true, video: false }, (stream) => {
            var audioContext = window.AudioContext;
            var context = new audioContext();
            var audioInput = context.createMediaStreamSource(stream);

            var recorder = context.createScriptProcessor(this.props.bufferSize, 1, 1);
            recorder.onaudioprocess = (e) => {
                var pcm = this.to16Bit(e.inputBuffer.getChannelData(0));
                this.props.onBufferCallback(pcm);
            }

            audioInput.connect(recorder);
            recorder.connect(context.destination);

            this.recorder = recorder;
            this.context = context;
            this.audioInput = audioInput;
        }, e => { console.log(e) });
    }

    stop() {
        this.audioInput.disconnect(this.recorder);
        this.recorder.disconnect(this.context.destination);
    }

    buttonClicked() {
        this.setState((prevState) => ({ recording: !prevState.recording }), () => {
            // Animate the button.
            $('.transform').toggleClass('transform-active');
            $('.icon').toggleClass('icon-active');

            this.props.onClick(this.state.recording);

            if (this.state.recording) {
                this.start();
            } else {
                this.stop();
            }
        });
    }

    to16Bit(buffer) {
        let l = buffer.length;
        let buf = new Int16Array(l);
        while (l--) {
            buf[l] = Math.min(1, buffer[l]) * 0x7FFF;
        }
        return buf;
    }

    render() {
        return (
            <div className="recorder transform" onClick={() => this.buttonClicked()}>
                <i className="icon material-icons">mic</i>
            </div>
        );
    }
}

BufferedRecorder.defaultProps = {
    bufferSize: 4096,
    sampleRate: 44100,
};

BufferedRecorder.propTypes = {
    onBufferCallback: PropTypes.func.isRequired
};

export default BufferedRecorder;