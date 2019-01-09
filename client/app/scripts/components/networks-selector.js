import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { selectNetwork, showNetworks } from '../actions/app-actions';
import { availableNetworksSelector } from '../selectors/node-networks';
import NetworkSelectorItem from './network-selector-item';

class NetworkSelector extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.onClick = this.onClick.bind(this);
    this.onMouseOut = this.onMouseOut.bind(this);
  }

  onClick() {
    return this.props.showNetworks(!this.props.showingNetworks);
  }

  onMouseOut() {
    this.props.selectNetwork(this.props.pinnedNetwork);
  }

  render() {
    const { availableNetworks, showingNetworks } = this.props;

    const items = availableNetworks.map(network => (
      <NetworkSelectorItem key={network.get('id')} network={network} />
    ));

    const className = classNames('network-selector-action', {
      'network-selector-action-selected': showingNetworks
    });

    const style = {
      borderBottomColor: showingNetworks ? '#A2A0B3' : 'transparent'
    };

    return (
      <div className="network-selector">
        <div className="network-selector-wrapper" onMouseLeave={this.onMouseOut}>
          <div className={className} onClick={this.onClick} style={style}>
            Networks
          </div>
          {showingNetworks && items}
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    availableNetworks: availableNetworksSelector(state),
    pinnedNetwork: state.get('pinnedNetwork'),
    showingNetworks: state.get('showingNetworks')
  };
}

export default connect(
  mapStateToProps,
  { selectNetwork, showNetworks }
)(NetworkSelector);
