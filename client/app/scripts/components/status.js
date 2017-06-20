import React from 'react';
import { connect } from 'react-redux';

import { isNowSelector } from '../selectors/time-travel';


class Status extends React.Component {
  render() {
    const { errorUrl, topologiesLoaded, filteredNodeCount, topology,
      websocketClosed, showingCurrentState } = this.props;

    let title = '';
    let text = 'Trying to reconnect...';
    let showWarningIcon = false;
    let classNames = 'status sidebar-item';

    if (errorUrl) {
      title = `Cannot reach Scope. Make sure the following URL is reachable: ${errorUrl}`;
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (!topologiesLoaded) {
      text = 'Connecting to Scope...';
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (websocketClosed) {
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (topology) {
      const stats = topology.get('stats');
      text = `${stats.get('node_count') - filteredNodeCount} nodes`;
      if (stats.get('filtered_nodes')) {
        text = `${text} (${stats.get('filtered_nodes') + filteredNodeCount} filtered)`;
      }
      classNames += ' status-stats';
      showWarningIcon = false;
      // TODO: Currently the stats are always pulled for the current state of the system,
      // so they are incorrect when showing the past. This should be addressed somehow.
      if (!showingCurrentState) {
        text = '';
      }
    }

    return (
      <div className={classNames}>
        {showWarningIcon && <span className="status-icon fa fa-exclamation-circle" />}
        <span className="status-label" title={title}>{text}</span>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    errorUrl: state.get('errorUrl'),
    filteredNodeCount: state.get('nodes').filter(node => node.get('filtered')).size,
    showingCurrentState: isNowSelector(state),
    topologiesLoaded: state.get('topologiesLoaded'),
    topology: state.get('currentTopology'),
    websocketClosed: state.get('websocketClosed'),
  };
}

export default connect(
  mapStateToProps
)(Status);
