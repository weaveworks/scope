import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import NodeResourcesMetricBox from './node-resources-metric-box';
import NodeResourcesLayerTopology from './node-resources-layer-topology';
import {
  layerVerticalPositionByTopologyIdSelector,
  layoutNodesByTopologyIdSelector,
} from '../../selectors/resource-view/layout';


class NodesResourcesLayer extends React.Component {
  render() {
    const {
      layerVerticalPosition, topologyId, transform, layoutNodes
    } = this.props;

    return (
      <g className="node-resources-layer">
        <g className="node-resources-metric-boxes">
          {layoutNodes.toIndexedSeq().map(node => (
            <NodeResourcesMetricBox
              id={node.get('id')}
              key={node.get('id')}
              color={node.get('color')}
              label={node.get('label')}
              topologyId={topologyId}
              metricSummary={node.get('metricSummary')}
              width={node.get('width')}
              height={node.get('height')}
              x={node.get('offset')}
              y={layerVerticalPosition}
              transform={transform}
            />
          ))}
        </g>
        {!layoutNodes.isEmpty() && <NodeResourcesLayerTopology
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
    layerVerticalPosition: layerVerticalPositionByTopologyIdSelector(state).get(props.topologyId),
    layoutNodes: layoutNodesByTopologyIdSelector(state).get(props.topologyId, makeMap()),
  };
}

export default connect(mapStateToProps)(NodesResourcesLayer);
