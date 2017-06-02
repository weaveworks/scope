import React from 'react';
import { connect } from 'react-redux';


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
      if (showingCurrentState) {
        text = `${stats.get('node_count') - filteredNodeCount} nodes`;
        if (stats.get('filtered_nodes')) {
          text = `${text} (${stats.get('filtered_nodes') + filteredNodeCount} filtered)`;
        }
      } else {
        text = '';
      }
      classNames += ' status-stats';
      showWarningIcon = false;
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
    showingCurrentState: !state.get('websocketQueryTimestamp'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topology: state.get('currentTopology'),
    websocketClosed: state.get('websocketClosed'),
  };
}

export default connect(
  mapStateToProps
)(Status);
