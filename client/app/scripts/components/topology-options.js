import React from 'react';

import TopologyOptionAction from './topology-option-action';

export default class TopologyOptions extends React.Component {

  renderOption(option) {
    const { activeOptions, topologyId } = this.props;
    const optionId = option.get('id');
    const activeValue = activeOptions && activeOptions.has(optionId)
      ? activeOptions.get(optionId) : option.get('defaultValue');

    return (
      <div className="topology-option" key={optionId}>
        <div className="topology-option-wrapper">
          {option.get('options').map(item => <TopologyOptionAction
            optionId={optionId} topologyId={topologyId} key={item.get('value')}
            activeValue={activeValue} item={item} />)}
        </div>
      </div>
    );
  }

  render() {
    return (
      <div className="topology-options">
        {this.props.options.toIndexedSeq().map(option => this.renderOption(option))}
      </div>
    );
  }
}
