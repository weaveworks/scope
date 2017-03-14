import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { RESOURCES_LAYER_TITLE_WIDTH, RESOURCES_LAYER_HEIGHT } from '../constants/styles';
import {
  layersVerticalPositionSelector,
  positionedNodesByTopologySelector,
} from '../selectors/resource-view/layers';
import NodeResourceBox from './node-resource-box';
import NodeResourceLabel from './node-resource-label';


const PADDING = 10;

const stringifiedTransform = ({ scaleX, scaleY, translateX, translateY }) => (
  `translate(${translateX},${translateY}) scale(${scaleX},${scaleY})`
);

const getPositionedLabels = (nodes, transform) => {
  const { scaleX, scaleY, translateX, translateY } = transform;
  return nodes.map((node) => {
    const nodeX = (node.get('x') * scaleX) + translateX;
    const nodeY = (node.get('y') * scaleY) + translateY;
    const nodeWidth = node.get('width') * scaleX;

    const labelY = nodeY + PADDING;
    const labelX = Math.max(200, nodeX) + PADDING;
    const labelWidth = (nodeX + nodeWidth) - PADDING - labelX;

    if (labelWidth < 20) return makeMap();

    return makeMap({
      id: node.get('id'),
      label: node.get('label'),
      width: labelWidth,
      x: labelX,
      y: labelY,
    });
  }).filter(label => !label.isEmpty());
};

class NodesResourcesLayer extends React.Component {
  render() {
    const { yPosition, topologyId, transform, nodes, labels } = this.props;
    const height = RESOURCES_LAYER_HEIGHT * transform.scaleY;

    return (
      <g className="node-resource-layer">
        <g transform={stringifiedTransform(transform)}>
          {nodes.toIndexedSeq().map(node => (
            <NodeResourceBox
              key={node.get('id')}
              color={node.get('color')}
              width={node.get('width')}
              height={node.get('height')}
              withCapacity={node.get('withCapacity')}
              x={node.get('x')}
              y={node.get('y')}
              activeMetric={node.get('activeMetric')}
              meta={node.get('meta')}
            />
          ))}
        </g>
        <g>
          {labels.toIndexedSeq().map(label => (
            <NodeResourceLabel
              key={label.get('id')}
              label={label.get('label')}
              width={label.get('width')}
              x={label.get('x')}
              y={label.get('y')}
            />
          ))}
        </g>
        {!nodes.isEmpty() && <foreignObject
          className="layer-name"
          y={(yPosition * transform.scaleY) + transform.translateY}
          height={height}
          width={RESOURCES_LAYER_TITLE_WIDTH}>
          <span style={{ height, lineHeight: `${height}px` }}>{topologyId}</span>
        </foreignObject>}
      </g>
    );
  }
}

function mapStateToProps(state, props) {
  const yPosition = layersVerticalPositionSelector(state).get(props.topologyId);
  const nodes = positionedNodesByTopologySelector(state).get(props.topologyId, makeMap());
  // TODO: Move to selectors?
  const labels = getPositionedLabels(nodes, props.transform);
  return { yPosition, nodes, labels };
}

export default connect(
  mapStateToProps
)(NodesResourcesLayer);
