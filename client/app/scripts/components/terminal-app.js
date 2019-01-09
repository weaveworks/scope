import React from 'react';
import { connect } from 'react-redux';

import Terminal from './terminal';
import { receiveControlPipeFromParams, hitEsc } from '../actions/app-actions';

const ESC_KEY_CODE = 27;

class TerminalApp extends React.Component {
  constructor(props, context) {
    super(props, context);

    const paramString = window.location.hash.split('/').pop();
    const params = JSON.parse(decodeURIComponent(paramString));
    this.props.receiveControlPipeFromParams(
      params.pipe.id, params.pipe.raw,
      params.pipe.resizeTtyControl
    );

    this.state = {
      statusBarColor: params.statusBarColor,
      title: params.title,
      titleBarColor: params.titleBarColor
    };

    this.onKeyUp = this.onKeyUp.bind(this);
  }

  componentDidMount() {
    window.addEventListener('keyup', this.onKeyUp);
  }

  componentWillUnmount() {
    window.removeEventListener('keyup', this.onKeyUp);
  }

  onKeyUp(ev) {
    if (ev.keyCode === ESC_KEY_CODE) {
      this.props.hitEsc();
    }
  }

  componentWillReceiveProps(nextProps) {
    if (!nextProps.controlPipe) {
      window.close();
    }
  }

  render() {
    const style = {borderTop: `4px solid ${this.state.titleBarColor}`};

    return (
      <div className="terminal-app" style={style}>
        {this.props.controlPipe && <Terminal
          pipe={this.props.controlPipe}
          titleBarColor={this.state.titleBarColor}
          statusBarColor={this.state.statusBarColor}
          title={this.state.title}
          embedded={false} />}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    controlPipe: state.get('controlPipes').last()
  };
}

export default connect(
  mapStateToProps,
  { hitEsc, receiveControlPipeFromParams }
)(TerminalApp);
