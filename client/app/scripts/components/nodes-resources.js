import React from 'react';
import { connect } from 'react-redux';

import Logo from './logo';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layers';
import CachableZoomWrapper from './cachable-zoom-wrapper';
import NodesResourcesLayer from './nodes-resources/node-resources-layer';


class NodesResources extends React.Component {
  renderLayers(transform) {
    return this.props.layersTopologyIds.map((topologyId, index) => (
      <NodesResourcesLayer
        key={topologyId}
        topologyId={topologyId}
        transform={transform}
        slot={index}
      />
    ));
  }

  render() {
    return (
      <div className="nodes-chart">
        <svg
          width="100%" height="100%"
          id="nodes-chart-canvas">
          <Logo transform="translate(24,24) scale(0.25)" />
          <CachableZoomWrapper bounded forwardTransform fixVertical>
            {transform => this.renderLayers(transform)}
          </CachableZoomWrapper>
        </svg>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    layersTopologyIds: layersTopologyIdsSelector(state),
  };
}

export default connect(
  mapStateToProps
)(NodesResources);
