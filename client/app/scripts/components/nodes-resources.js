import React from 'react';
import { connect } from 'react-redux';

import ZoomableCanvas from './zoomable-canvas';
import NodesResourcesLayer from './nodes-resources/node-resources-layer';
import { layersTopologyIdsSelector } from '../selectors/resource-view/layout';
import {
  resourcesLimitsSelector,
  resourcesZoomStateSelector,
} from '../selectors/resource-view/zoom';
import { clickBackground } from '../actions/app-actions';

import { CONTENT_COVERING } from '../constants/naming';


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
          onClick={this.handleMouseClick}
          fixVertical
          boundContent={CONTENT_COVERING}
          limitsSelector={resourcesLimitsSelector}
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
    selectedNodeId: state.get('selectedNodeId'),
  };
}

export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesResources);
