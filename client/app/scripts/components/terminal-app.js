import React from 'react';

import AppStore from '../stores/app-store';
import Terminal from './terminal';
import { receiveControlPipeFromParams } from '../actions/app-actions';

function getStateFromStores() {
  return {
    controlPipe: AppStore.getControlPipe()
  };
}

export class TerminalApp extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onChange = this.onChange.bind(this);

    const paramString = window.location.hash.split('/').pop();
    const params = JSON.parse(decodeURIComponent(paramString));
    receiveControlPipeFromParams(params.pipe.id, null, params.pipe.raw, false);

    this.state = {
      title: params.title,
      titleBarColor: params.titleBarColor,
      statusBarColor: params.statusBarColor,
      controlPipe: AppStore.getControlPipe()
    };
  }

  componentDidMount() {
    AppStore.addListener(this.onChange);
  }

  onChange() {
    this.setState(getStateFromStores());
  }

  render() {
    const style = {borderTop: `4px solid ${this.state.titleBarColor}`};

    return (
      <div className="terminal-app" style={style}>
        {this.state.controlPipe && <Terminal
          pipe={this.state.controlPipe}
          titleBarColor={this.state.titleBarColor}
          statusBarColor={this.state.statusBarColor}
          title={this.state.title}
          embedded={false} />}
      </div>
    );
  }
}
