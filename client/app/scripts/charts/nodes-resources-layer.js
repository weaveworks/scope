import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import NodeResourceBox from './node-resource-box';
import LayerLabelsOverlay from './layer-labels-overlay';
import LayerTopologyName from './layer-topology-name';
import {
  layersVerticalPositionSelector,
  positionedNodesByTopologySelector,
} from '../selectors/resource-view/layers';


// const stringifiedTransform = ({ scaleX = 1, scaleY = 1, translateX = 0, translateY = 0 }) => (
//   `translate(${translateX},${translateY}) scale(${scaleX},${scaleY})`
// );

class NodesResourcesLayer extends React.Component {
  render() {
    const { layerVerticalPosition, topologyId, transform, nodes } = this.props;

    return (
      <g className="node-resource-layer">
        <g>
          {nodes.toIndexedSeq().map(node => (
            <NodeResourceBox
              key={node.get('id')}
              color={node.get('color')}
              width={node.get('width')}
              height={node.get('height')}
              withCapacity={node.get('withCapacity')}
              activeMetric={node.get('activeMetric')}
              x={node.get('offset')}
              y={layerVerticalPosition}
              transform={transform}
            />
          ))}
        </g>
        <LayerLabelsOverlay
          verticalOffset={layerVerticalPosition}
          transform={transform}
          nodes={nodes}
        />
        {!nodes.isEmpty() && <LayerTopologyName
          verticalOffset={layerVerticalPosition}
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
    nodes: positionedNodesByTopologySelector(state).get(props.topologyId, makeMap())
  };
}

export default connect(
  mapStateToProps
)(NodesResourcesLayer);
