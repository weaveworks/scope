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

  render: function() {
    return (
      <div className="status">
        {this.renderConnectionState(this.props.errorUrl, this.props.websocketClosed)}
      </div>
    );
  }

});

module.exports = Status;
