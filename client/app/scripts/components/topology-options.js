import React from 'react';
import { connect } from 'react-redux';
import { Set as makeSet, Map as makeMap } from 'immutable';
import includes from 'lodash/includes';

import { trackAnalyticsEvent } from '../utils/tracking-utils';
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
    trackAnalyticsEvent('scope.topology.option.click', {
      layout: this.props.topologyViewMode,
      optionId,
      parentTopologyId: this.props.currentTopology.get('parentId'),
      topologyId: this.props.currentTopology.get('id'),
      value: nextOptions,
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
      const selectedActiveOptions = opts[selected] || [];
      const isSelectedAlready = includes(selectedActiveOptions, value);

      if (isSelectedAlready) {
        // Remove the option if it is already selected
        nextOptions = selectedActiveOptions.filter(o => o !== value);
      } else {
        // Add it to the array if it's not selected
        nextOptions = selectedActiveOptions.concat(value);
      }
      // Since the user is clicking an option, remove the highlighting from the none option,
      // unless they are removing the last option. In that case, default to the none label.
      // Note that since the other ids are potentially user-controlled (eg. k8s namespaces),
      // the only string we can use for the none option is the empty string '',
      // since that can't collide.
      if (nextOptions.length === 0) {
        nextOptions = [''];
      } else {
        nextOptions = nextOptions.filter(o => o !== '');
      }
    }
    this.trackOptionClick(optionId, nextOptions);
    this.props.changeTopologyOption(optionId, nextOptions, topologyId);
  }

  handleNoneClick(optionId, value, topologyId) {
    const nextOptions = [''];
    this.trackOptionClick(optionId, nextOptions);
    this.props.changeTopologyOption(optionId, nextOptions, topologyId);
  }

  renderOption(option) {
    const { activeOptions, currentTopologyId } = this.props;
    const optionId = option.get('id');

    // Make the active value be the intersection of the available options
    // and the active selection and use the default value if there is no
    // overlap. It seems intuitive that active selection would always be a
    // subset of available option, but the exception can happen when going
    // back in time (making available options change, while not touching
    // the selection).
    // TODO: This logic should probably be made consistent with how topology
    // selection is handled when time travelling, especially when the name-
    // spaces are brought under category selection.
    // TODO: Consider extracting this into a global selector.
    let activeValue = option.get('defaultValue');
    if (activeOptions && activeOptions.has(optionId)) {
      const activeSelection = makeSet(activeOptions.get(optionId));
      const availableOptions = makeSet(option.get('options').map(o => o.get('value')));
      const intersection = activeSelection.intersect(availableOptions);
      if (!intersection.isEmpty()) {
        activeValue = intersection.toJS();
      }
    }

    const noneItem = makeMap({
      label: option.get('noneLabel'),
      value: ''
    });
    return (
      <div className="topology-option" key={optionId}>
        <div className="topology-option-wrapper">
          {option.get('selectType') === 'union' &&
            <TopologyOptionAction
              onClick={this.handleNoneClick}
              optionId={optionId}
              item={noneItem}
              topologyId={currentTopologyId}
              activeValue={activeValue}
            />
          }
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
        </div>
      </div>
    );
  }

  render() {
    const { options } = this.props;
    return (
      <div className="topology-options">
        {options && options.toIndexedSeq().map(option => this.renderOption(option))}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    activeOptions: activeTopologyOptionsSelector(state),
    currentTopology: state.get('currentTopology'),
    currentTopologyId: state.get('currentTopologyId'),
    options: getCurrentTopologyOptions(state),
    topologyViewMode: state.get('topologyViewMode')
  };
}

export default connect(
  mapStateToProps,
  { changeTopologyOption }
)(TopologyOptions);
