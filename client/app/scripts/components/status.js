const React = require('react');

const Status = React.createClass({

  renderConnectionState: function() {
    return (
      <div className="status-connection">
        <span className="status-icon fa fa-exclamation-circle" />
        <span className="status-label">Scope is disconnected</span>
      </div>
    );
  },

  render: function() {
    const isDisconnected = this.props.connectionState === 'disconnected';

    return (
      <div className="status">
        {isDisconnected && this.renderConnectionState()}
      </div>
    );
  }

});

module.exports = Status;
