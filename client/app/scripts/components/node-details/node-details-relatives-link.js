import React from 'react';

import { clickRelative } from '../../actions/app-actions';
import MatchedText from '../matched-text';

export default class NodeDetailsRelativesLink extends React.PureComponent {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
    this.saveNodeRef = this.saveNodeRef.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.dispatch(clickRelative(
      this.props.id,
      this.props.topologyId,
      this.props.label,
      this.node.getBoundingClientRect()
    ));
  }

  saveNodeRef(ref) {
    this.node = ref;
  }

  render() {
    const title = `View in ${this.props.topologyId}: ${this.props.label}`;
    return (
      <span
        className="node-details-relatives-link" title={title}
        onClick={this.handleClick} ref={this.saveNodeRef}>
        <MatchedText text={this.props.label} match={this.props.match} />
      </span>
    );
  }
}
