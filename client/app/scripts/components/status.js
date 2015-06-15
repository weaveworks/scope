const React = require('react');

const Status = React.createClass({

  renderConnectionState: function(errorUrl) {
    if (errorUrl) {
      const title = 'Cannot reach Scope. Make sure the following URL is reachable: ' + errorUrl;
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
        {this.renderConnectionState(this.props.errorUrl)}
      </div>
    );
  }

});

module.exports = Status;
