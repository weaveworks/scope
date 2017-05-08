import React from 'react';
import { connect } from 'react-redux';

import ZoomableCanvas from './zoomable-canvas';
import NodesResourcesLayer from './nodes-resources/node-resources-layer';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import {
  resourcesZoomLimitsSelector,
  resourcesZoomStateSelector,
} from '../selectors/resource-view/zoom';


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
      <div className="nodes-resources">
        <ZoomableCanvas
          bounded forwardTransform fixVertical
          zoomLimitsSelector={resourcesZoomLimitsSelector}
          zoomStateSelector={resourcesZoomStateSelector}>
          {transform => this.renderLayers(transform)}
        </ZoomableCanvas>
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
