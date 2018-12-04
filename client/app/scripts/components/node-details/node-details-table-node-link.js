import React from 'react';
import { connect } from 'react-redux';

import { clickRelative } from '../../actions/app-actions';
import { trackAnalyticsEvent } from '../../utils/tracking-utils';
import { dismissRowClickProps } from './node-details-table-row';


class NodeDetailsTableNodeLink extends React.Component {
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
      this.props.nodeId,
      this.props.topologyId,
      this.props.label,
      this.node.getBoundingClientRect()
    ));
  }

  saveNodeRef(ref) {
    this.node = ref;
  }

  render() {
    const { label, labelMinor } = this.props;
    const title = !labelMinor ? label : `${label} (${labelMinor})`;

    return (
      <span
        className="node-details-table-node-link"
        title={title}
        ref={this.saveNodeRef}
        onClick={this.handleClick}
        {...dismissRowClickProps}
      >
        {label}
      </span>
    );
  }
}

// Using this instead of PureComponent because of props.dispatch
export default connect()(NodeDetailsTableNodeLink);
