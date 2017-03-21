import React from 'react';
import { connect } from 'react-redux';

import { getCurrentTopologyOptions } from '../utils/topology-utils';
import { activeTopologyOptionsSelector } from '../selectors/topology';
import TopologyOptionAction from './topology-option-action';

class TopologyOptions extends React.Component {

  renderOption(option) {
    const { activeOptions, topologyId } = this.props;
    const optionId = option.get('id');
    const activeValue = activeOptions && activeOptions.has(optionId)
      ? activeOptions.get(optionId)
      : option.get('defaultValue');

    return (
      <div className="topology-option" key={optionId}>
        <div className="topology-option-wrapper">
          {option.get('options').map(item => (
            <TopologyOptionAction
              optionId={optionId}
              topologyId={topologyId}
              key={item.get('value')}
              activeValue={activeValue}
              item={item}
            />
          ))}
        </div>
      </div>
    );
  }

  render() {
    return (
      <div className="topology-options">
        {this.props.options && this.props.options.toIndexedSeq().map(
          option => this.renderOption(option))}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    options: getCurrentTopologyOptions(state),
    topologyId: state.get('currentTopologyId'),
    activeOptions: activeTopologyOptionsSelector(state)
  };
}

export default connect(
  mapStateToProps
)(TopologyOptions);
