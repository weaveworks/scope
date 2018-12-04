import React from 'react';
import { connect } from 'react-redux';

import { clickRelative } from '../../actions/app-actions';
import { trackAnalyticsEvent } from '../../utils/tracking-utils';
import MatchedText from '../matched-text';


class NodeDetailsRelativesLink extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleClick = this.handleClick.bind(this);
    this.saveNodeRef = this.saveNodeRef.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    trackAnalyticsEvent('scope.node.relative.click', {
      topologyId: this.props.topologyId,
    });
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
        className="node-details-relatives-link"
        title={title}
        onClick={this.handleClick}
        ref={this.saveNodeRef}>
        <MatchedText text={this.props.label} match={this.props.match} />
      </span>
    );
  }
}

// Using this instead of PureComponent because of props.dispatch
export default connect()(NodeDetailsRelativesLink);
