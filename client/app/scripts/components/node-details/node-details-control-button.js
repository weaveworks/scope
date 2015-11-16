const React = require('react');

const AppActions = require('../../actions/app-actions');

const NodeDetailsControlButton = React.createClass({

  render: function() {
    let className = `node-control-button fa ${this.props.control.icon}`;
    if (this.props.pending) {
      className += ' node-control-button-pending';
    }
    return (
      <span className={className} title={this.props.control.human} onClick={this.handleClick} />
    );
  },

  handleClick: function(ev) {
    ev.preventDefault();
    AppActions.doControl(this.props.control.probeId, this.props.control.nodeId, this.props.control.id);
  }

});

module.exports = NodeDetailsControlButton;
