import React from 'react';

export default function Status({errorUrl, topologiesLoaded, topology, websocketClosed}) {
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
