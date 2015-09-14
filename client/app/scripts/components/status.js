const React = require('react');

const Status = React.createClass({

  renderConnectionState: function(errorUrl, websocketClosed) {
    if (errorUrl || websocketClosed) {
      const title = errorUrl ? 'Cannot reach Scope. Make sure the following URL is reachable: ' + errorUrl : '';
      return (
        <div className="status-connection" title={title}>
          <span className="status-icon fa fa-exclamation-circle" />
          <span className="status-label">Trying to reconnect...</span>
        </div>
      );
    }
  },

  renderTopologyStats: function(stats) {
    const statsText = `${stats.node_count} nodes, ${stats.edge_count} connections`;
    return <div className="status-stats">{statsText}</div>;
  },

  render: function() {
    const showStats = this.props.topology && !this.props.errorUrl && !this.props.websocketClosed;
    return (
      <div className="status sidebar-item">
        {showStats && this.renderTopologyStats(this.props.topology.stats)}
        {!showStats && this.renderConnectionState(this.props.errorUrl, this.props.websocketClosed)}
      </div>
    );
  }

});

module.exports = Status;
