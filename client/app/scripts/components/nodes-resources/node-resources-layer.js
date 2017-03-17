import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import NodeResourcesMetricBox from './node-resources-metric-box';
import NodeResourcesLayerTopology from './node-resources-layer-topology';
import {
  layersVerticalPositionSelector,
  positionedNodesByTopologySelector,
} from '../../selectors/resource-view/layers';


class NodesResourcesLayer extends React.Component {
  render() {
    const { layerVerticalPosition, topologyId, transform, nodes } = this.props;

    return (
      <g className="node-resources-layer">
        <g className="node-resources-metric-boxes">
          {nodes.toIndexedSeq().map(node => (
            <NodeResourcesMetricBox
              key={node.get('id')}
              color={node.get('color')}
              label={node.get('label')}
              withCapacity={node.get('withCapacity')}
              activeMetric={node.get('activeMetric')}
              width={node.get('width')}
              height={node.get('height')}
              x={node.get('offset')}
              y={layerVerticalPosition}
              transform={transform}
            />
          ))}
        </g>
        {!nodes.isEmpty() && <NodeResourcesLayerTopology
          verticalPosition={layerVerticalPosition}
          transform={transform}
          topologyId={topologyId}
        />}
      </g>
    );
  }
}

function mapStateToProps(state, props) {
  return {
    layerVerticalPosition: layersVerticalPositionSelector(state).get(props.topologyId),
    nodes: positionedNodesByTopologySelector(state).get(props.topologyId, makeMap()),
  };
}

export default connect(
  mapStateToProps
)(NodesResourcesLayer);
