import React from 'react';
import ReactDOM from 'react-dom';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsTableNodeLink extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    clickRelative(this.props.id, this.props.topologyId, this.props.label,
      ReactDOM.findDOMNode(this).getBoundingClientRect());
  }

  render() {
    if (this.props.linkable) {
      return (
        <span className="node-details-table-node-link truncate" title={this.props.label}
          onClick={this.handleClick}>
          {this.props.label}
        </span>
      );
    }
    return (
      <span className="node-details-table-node truncate" title={this.props.label}>
        {this.props.label}
      </span>
    );
  }
}

reactMixin.onClass(NodeDetailsTableNodeLink, PureRenderMixin);
