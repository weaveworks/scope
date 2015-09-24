const React = require('react');

const AppActions = require('../actions/app-actions');

const TopologyOptionAction = React.createClass({

  onClick: function(ev) {
    ev.preventDefault();
    AppActions.changeTopologyOption(this.props.option, this.props.value, this.props.topologyId);
  },

  render: function() {
    return (
      <span className="sidebar-item-action" onClick={this.onClick}>
        {this.props.value}
      </span>
    );
  }

});

module.exports = TopologyOptionAction;
