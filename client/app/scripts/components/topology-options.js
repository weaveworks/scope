import React from 'react';
import _ from 'lodash';

import TopologyOptionAction from './topology-option-action';

export default class TopologyOptions extends React.Component {
  renderAction(action, option, topologyId) {
    return (
      <TopologyOptionAction option={option} value={action} topologyId={topologyId} key={action} />
    );
  }

  /**
   * transforms a list of options into one sidebar-item.
   * The sidebar text comes from the active option. the actions come from the
   * remaining items.
   */
  renderOption(items) {
    let activeText;
    let activeValue;
    const actions = [];
    const activeOptions = this.props.activeOptions;
    const topologyId = this.props.topologyId;
    const option = items[0].option;

    // find active option value
    if (activeOptions && activeOptions.has(option)) {
      activeValue = activeOptions.get(option);
    } else {
      // get default value
      items.forEach(function(item) {
        if (item.default) {
          activeValue = item.value;
        }
      });
    }

    // render active option as text, add other options as actions
    items.forEach(function(item) {
      if (item.value === activeValue) {
        activeText = item.display;
      } else {
        actions.push(this.renderAction(item.value, item.option, topologyId));
      }
    }, this);

    return (
      <div className="sidebar-item" key={option}>
        {activeText}
        <span className="sidebar-item-actions">
          {actions}
        </span>
      </div>
    );
  }

  render() {
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
}
