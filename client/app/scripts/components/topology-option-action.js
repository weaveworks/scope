import React from 'react';

import { changeTopologyOption } from '../actions/app-actions';

export default class TopologyOptionAction extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.onClick = this.onClick.bind(this);
  }

  onClick(ev) {
    ev.preventDefault();
    changeTopologyOption(this.props.option, this.props.value, this.props.topologyId);
  }

  render() {
    return (
      <span className="sidebar-item-action" onClick={this.onClick}>
        {this.props.value}
      </span>
    );
  }
}
