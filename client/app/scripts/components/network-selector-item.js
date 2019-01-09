import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { selectNetwork, pinNetwork, unpinNetwork } from '../actions/app-actions';
import { getNetworkColor } from '../utils/color-utils';

class NetworkSelectorItem extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.onMouseOver = this.onMouseOver.bind(this);
    this.onMouseClick = this.onMouseClick.bind(this);
  }

  onMouseOver() {
    const k = this.props.network.get('id');
    this.props.selectNetwork(k);
  }

  onMouseClick() {
    const k = this.props.network.get('id');
    const { pinnedNetwork } = this.props;

    if (k === pinnedNetwork) {
      this.props.unpinNetwork(k);
    } else {
      this.props.pinNetwork(k);
    }
  }

  render() {
    const {network, selectedNetwork, pinnedNetwork} = this.props;
    const id = network.get('id');
    const isPinned = (id === pinnedNetwork);
    const isSelected = (id === selectedNetwork);
    const className = classNames('network-selector-action', {
      'network-selector-action-selected': isSelected
    });
    const style = {
      borderBottomColor: getNetworkColor(network.get('colorKey', id))
    };

    return (
      <div
        key={id}
        className={className}
        onMouseOver={this.onMouseOver}
        onClick={this.onMouseClick}
        style={style}>
        {network.get('label')}
        {isPinned && <i className="fa fa-thumbtack" />}
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    pinnedNetwork: state.get('pinnedNetwork'),
    selectedNetwork: state.get('selectedNetwork')
  };
}

export default connect(
  mapStateToProps,
  { pinNetwork, selectNetwork, unpinNetwork }
)(NetworkSelectorItem);
