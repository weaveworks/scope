import React from 'react';

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
    const option = items.first().get('option');

    // find active option value
    if (activeOptions && activeOptions.has(option)) {
      activeValue = activeOptions.get(option);
    } else {
      // get default value
      items.forEach(item => {
        if (item.get('default')) {
          activeValue = item.get('value');
        }
      });
    }

    // render active option as text, add other options as actions
    items.forEach(item => {
      if (item.get('value') === activeValue) {
        activeText = item.get('display');
      } else {
        actions.push(this.renderAction(item.get('value'), item.get('option'), topologyId));
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
    const options = this.props.options.map((items, optionId) => {
      let itemsMap = items.map(item => item.set('option', optionId));
      itemsMap = itemsMap.set('option', optionId);
      return itemsMap;
    });

    return (
      <div className="topology-options">
        {options.toIndexedSeq().map(items => this.renderOption(items))}
      </div>
    );
  }
}
