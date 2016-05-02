import React from 'react';
import { connect } from 'react-redux';

import { changeTopologyOption } from '../actions/app-actions';

class TopologyOptionAction extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onClick = this.onClick.bind(this);
  }

  onClick(ev) {
    ev.preventDefault();
    const { optionId, topologyId, item } = this.props;
    this.props.changeTopologyOption(optionId, item.get('value'), topologyId);
  }

  render() {
    const { activeValue, item } = this.props;
    const className = activeValue === item.get('value')
      ? 'topology-option-action topology-option-action-selected' : 'topology-option-action';
    return (
      <div className={className} onClick={this.onClick}>
        {item.get('label')}
      </div>
    );
  }
}

export default connect(
  null,
  { changeTopologyOption }
)(TopologyOptionAction);
