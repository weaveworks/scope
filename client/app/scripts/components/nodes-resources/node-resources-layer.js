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
    const { layerVerticalPosition, topologyId, transform, layoutNodes } = this.props;
    let height;
    if (topologyId === 'hosts') height = 400;
    if (topologyId === 'containers') height = 335;
    if (topologyId === 'processes') height = 270;

    let y;
    if (topologyId === 'hosts') y = -400;
    if (topologyId === 'containers') y = -345;
    if (topologyId === 'processes') y = -290;

    return (
      <g className="node-resources-layer">
        <g className="node-resources-metric-boxes">
          {layoutNodes.toIndexedSeq().map(node => (
            <NodeResourcesMetricBox
              key={node.get('id')}
              color={node.get('color')}
              label={node.get('label')}
              metricSummary={node.get('metricSummary')}
              width={node.get('width')}
              height={height}
              x={node.get('offset')}
              topId={topologyId}
              y={y}
              transform={transform}
            />
          ))}
        </g>
        {!layoutNodes.isEmpty() && false && <NodeResourcesLayerTopology
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
    layerVerticalPosition: layerVerticalPositionByTopologyIdSelector(state).get('hosts'),
    layoutNodes: layoutNodesByTopologyIdSelector(state).get(props.topologyId, makeMap()),
  };
}

export default connect(
  mapStateToProps
)(NodesResourcesLayer);
