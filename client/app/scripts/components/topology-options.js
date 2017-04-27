import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';
import includes from 'lodash/includes';

import { trackMixpanelEvent } from '../utils/tracking-utils';
import { getCurrentTopologyOptions } from '../utils/topology-utils';
import { activeTopologyOptionsSelector } from '../selectors/topology';
import TopologyOptionAction from './topology-option-action';
import { changeTopologyOption } from '../actions/app-actions';

class TopologyOptions extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.trackOptionClick = this.trackOptionClick.bind(this);
    this.handleOptionClick = this.handleOptionClick.bind(this);
    this.handleNoneClick = this.handleNoneClick.bind(this);
  }

  trackOptionClick(optionId, nextOptions) {
    trackMixpanelEvent('scope.topology.option.click', {
      optionId,
      value: nextOptions,
      layout: this.props.topologyViewMode,
      topologyId: this.props.currentTopology.get('id'),
      parentTopologyId: this.props.currentTopology.get('parentId'),
    });
  }

  handleOptionClick(optionId, value, topologyId) {
    let nextOptions = [value];
    const { activeOptions, options } = this.props;
    const selectedOption = options.find(o => o.get('id') === optionId);

    if (selectedOption.get('selectType') === 'union') {
      // Multi-select topology options (such as k8s namespaces) are handled here.
      // Users can select one, many, or none of these options.
      // The component builds an array of the next selected values that are sent to the action.
      const opts = activeOptions.toJS();
      const selected = selectedOption.get('id');
      const isSelectedAlready = includes(opts[selected], value);

      if (isSelectedAlready) {
        // Remove the option if it is already selected
        nextOptions = opts[selected].filter(o => o !== value);
      } else {
        // Add it to the array if it's not selected
        nextOptions = opts[selected].concat(value);
      }
      // Since the user is clicking an option, remove the highlighting from the 'none' option,
      // unless they are removing the last option. In that case, default to the 'none' label.
      if (nextOptions.length === 0) {
        nextOptions = ['none'];
      } else {
        nextOptions = nextOptions.filter(o => o !== 'none');
      }
    }
    this.trackOptionClick(optionId, nextOptions);
    this.props.changeTopologyOption(optionId, nextOptions, topologyId);
  }

  handleNoneClick(optionId, value, topologyId) {
    const nextOptions = ['none'];
    this.trackOptionClick(optionId, nextOptions);
    this.props.changeTopologyOption(optionId, nextOptions, topologyId);
  }

  renderOption(option) {
    const { activeOptions, currentTopologyId } = this.props;
    const optionId = option.get('id');
    const activeValue = activeOptions && activeOptions.has(optionId)
      ? activeOptions.get(optionId)
      : option.get('defaultValue');
    const noneItem = makeMap({
      value: 'none',
      label: option.get('noneLabel')
    });
    return (
      <div className="topology-option" key={optionId}>
        <div className="topology-option-wrapper">
          {option.get('options').map(item => (
            <TopologyOptionAction
              onClick={this.handleOptionClick}
              optionId={optionId}
              topologyId={currentTopologyId}
              key={item.get('value')}
              activeValue={activeValue}
              item={item}
            />
          ))}
          {option.get('selectType') === 'union' &&
            <TopologyOptionAction
              onClick={this.handleNoneClick}
              optionId={optionId}
              item={noneItem}
              topologyId={currentTopologyId}
              activeValue={activeValue}
            />
          }
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
    topologyViewMode: state.get('topologyViewMode'),
    currentTopology: state.get('currentTopology'),
    currentTopologyId: state.get('currentTopologyId'),
    activeOptions: activeTopologyOptionsSelector(state)
  };
}

export default connect(
  mapStateToProps,
  { changeTopologyOption }
)(TopologyOptions);
