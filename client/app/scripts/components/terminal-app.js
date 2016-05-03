import React from 'react';
import { connect } from 'react-redux';

import Terminal from './terminal';
import { receiveControlPipeFromParams } from '../actions/app-actions';

class TerminalApp extends React.Component {

  constructor(props, context) {
    super(props, context);

    const paramString = window.location.hash.split('/').pop();
    const params = JSON.parse(decodeURIComponent(paramString));
    this.props.receiveControlPipeFromParams(params.pipe.id, null, params.pipe.raw, false);

    this.state = {
      title: params.title,
      titleBarColor: params.titleBarColor,
      statusBarColor: params.statusBarColor
    };
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
  { receiveControlPipeFromParams }
)(TerminalApp);
