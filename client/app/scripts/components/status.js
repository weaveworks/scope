import React from 'react';
import { connect } from 'react-redux';

class Status extends React.Component {
  render() {
    const {errorUrl, topology, websocketClosed} = this.props;

    let title = '';
    let text = 'Trying to reconnect...';
    let showWarningIcon = false;
    let classNames = 'status sidebar-item';

    if (errorUrl) {
      title = `Cannot reach Scope. Make sure the following URL is reachable: ${errorUrl}`;
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (websocketClosed) {
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (topology) {
      const stats = topology.get('stats');
      text = `${stats.get('node_count')} nodes`;
      if (stats.get('filtered_nodes')) {
        text = `${text} (${stats.get('filtered_nodes')} filtered)`;
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
    topologiesLoaded: state.get('topologiesLoaded'),
    topology: state.get('currentTopology'),
    websocketClosed: state.get('websocketClosed')
  };
}

export default connect(
  mapStateToProps
)(Status);
