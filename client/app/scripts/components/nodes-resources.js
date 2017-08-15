import React from 'react';
import { connect } from 'react-redux';

import ZoomableCanvas from './zoomable-canvas';
import NodesResourcesLayer from './nodes-resources/node-resources-layer';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import {
  resourcesZoomLimitsSelector,
  resourcesZoomStateSelector,
} from '../selectors/resource-view/zoom';
import { clickBackground } from '../actions/app-actions';


class NodesResources extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  handleMouseClick() {
    if (this.props.selectedNodeId) {
      this.props.clickBackground();
    }
  }

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
          bounded fixVertical
          onClick={this.handleMouseClick}
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
    selectedNodeId: state.get('selectedNodeId'),
    layersTopologyIds: layersTopologyIdsSelector(state),
  };
}

export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesResources);
