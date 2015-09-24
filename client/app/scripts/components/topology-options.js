const React = require('react');
const _ = require('lodash');

const TopologyOptionAction = require('./topology-option-action');

const TopologyOptions = React.createClass({

  renderAction: function(action, option, topologyId) {
    return (
      <TopologyOptionAction option={option} value={action} topologyId={topologyId} />
    );
  },

  renderOption: function(items) {
    let activeText;
    const actions = [];
    const activeOptions = this.props.activeOptions;
    const topologyId = this.props.topologyId;
    items.forEach(function(item) {
      if (activeOptions && activeOptions.has(item.option) && activeOptions.get(item.option) === item.value) {
        activeText = item.display;
      } else {
        actions.push(this.renderAction(item.value, item.option, topologyId));
      }
    }, this);

    return (
      <div className="sidebar-item">
        {activeText}
        <span className="sidebar-item-actions">
          {actions}
        </span>
      </div>
    );
  },

  render: function() {
    const options = _.sortBy(
      _.map(this.props.options, function(items, optionId) {
        _.each(items, function(item) {
          item.option = optionId;
        });
        items.option = optionId;
        return items;
      }),
      'option'
    );

    return (
      <div className="topology-options">
        {options.map(function(items) {
          return this.renderOption(items);
        }, this)}
      </div>
    );
  }

});

module.exports = TopologyOptions;
