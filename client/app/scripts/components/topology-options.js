import React from 'react';
import { connect } from 'react-redux';

import { getCurrentTopologyOptions } from '../utils/topology-utils';
import { activeTopologyOptionsSelector } from '../selectors/topology';
import TopologyOptionAction from './topology-option-action';
import { changeTopologyOption } from '../actions/app-actions';

class TopologyOptions extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleOptionClick = this.handleOptionClick.bind(this);
  }

  handleOptionClick(optionId, value, topologyId) {
    const { activeOptions, options } = this.props;
    const selectedOption = options.find(o => o.get('id') === optionId);

    if (selectedOption.get('selectType') === 'union') {
      const isSelectedAlready = activeOptions.get(selectedOption.get('id')).includes(value);
      const addOrRemove = isSelectedAlready ? 'remove' : 'add';
      this.props.changeTopologyOption(optionId, value, topologyId, addOrRemove);
    } else {
      this.props.changeTopologyOption(optionId, value, topologyId);
    }
  }

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
              onClick={this.handleOptionClick}
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
    const { options } = this.props;
    return (
      <div className="topology-options">
        {options && options.toIndexedSeq().map(
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
  mapStateToProps,
  { changeTopologyOption }
)(TopologyOptions);
