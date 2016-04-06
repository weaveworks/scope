import React from 'react';
import ReactDOM from 'react-dom';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import { clickRelative } from '../../actions/app-actions';

export default class NodeDetailsRelativesLink extends React.Component {

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
    const title = `View in ${this.props.topologyId}: ${this.props.label}`;
    return (
      <span className="node-details-relatives-link" title={title} onClick={this.handleClick}>
        {this.props.label}
      </span>
    );
  }
}

reactMixin.onClass(NodeDetailsRelativesLink, PureRenderMixin);
