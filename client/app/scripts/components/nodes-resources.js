import React from 'react';
import { connect } from 'react-redux';

import Logo from './logo';
import ZoomWrapper from './zoom-wrapper';
import NodeResourcesLayer from './nodes-resources/node-resources-layer';
import NodeResourcesZoomScale from './nodes-resources/node-resources-zoom-scale';
import {
  layersTopologyIdsSelector,
  layoutNodesByTopologyIdSelector,
} from '../selectors/resource-view/layout';
import {
  resourcesZoomLimitsSelector,
  resourcesZoomStateSelector,
} from '../selectors/resource-view/zoom';


class NodesResources extends React.Component {
  render() {
    const { layersTopologyIds, hasNodes } = this.props;

    return (
      <div className="nodes-resources">
        <svg id="canvas" width="100%" height="100%">
          <Logo transform="translate(24,24) scale(0.25)" />
          <ZoomWrapper
            svg="canvas" bounded forwardTransform fixVertical
            zoomLimitsSelector={resourcesZoomLimitsSelector}
            zoomStateSelector={resourcesZoomStateSelector}>
            {transform => hasNodes && (
              <g className="nodes-resources-zoomed">
                <g className="nodes-resources-layers">
                  {layersTopologyIds.map((topologyId, index) => (
                    <NodeResourcesLayer
                      key={topologyId}
                      topologyId={topologyId}
                      transform={transform}
                      slot={index}
                    />
                  ))}
                </g>
                <NodeResourcesZoomScale zoomLevel={transform.scaleX} />
              </g>
            )}
          </ZoomWrapper>
        </svg>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    layersTopologyIds: layersTopologyIdsSelector(state),
    hasNodes: !layoutNodesByTopologyIdSelector(state).isEmpty(),
  };
}

export default connect(
  mapStateToProps
)(NodesResources);
