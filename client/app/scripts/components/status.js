const React = require('react');

const Status = React.createClass({

  render: function() {
    let title = '';
    let text = 'Trying to reconnect...';
    let showWarningIcon = false;
    let classNames = 'status sidebar-item';

    if (this.props.errorUrl) {
      title = `Cannot reach Scope. Make sure the following URL is reachable: ${this.props.errorUrl}`;
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (!this.props.topologiesLoaded) {
      text = 'Loading topologies...';
      classNames += ' status-loading';
      showWarningIcon = false;
    } else if (this.props.websocketClosed) {
      classNames += ' status-loading';
      showWarningIcon = true;
    } else if (this.props.topology) {
      const stats = this.props.topology.stats;
      text = `${stats.node_count} nodes`;
      if (stats.filtered_nodes) {
        text = `${text} (${stats.filtered_nodes} filtered)`;
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

});

module.exports = Status;
