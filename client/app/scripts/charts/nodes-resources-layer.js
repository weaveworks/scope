import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { layoutNodesByTopologyMetaSelector } from '../selectors/resource-view/layout';
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
    const labelX = Math.max(0, nodeX) + PADDING;
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
    const { transform, nodes, labels } = this.props;
    return (
      <g>
        <g transform={stringifiedTransform(transform)}>
          {nodes.map(node => (
            <NodeResourceBox
              key={node.get('id')}
              color={node.get('color')}
              width={node.get('width')}
              height={node.get('height')}
              consumption={node.get('consumption')}
              withCapacity={node.get('withCapacity')}
              x={node.get('x')}
              y={node.get('y')}
              info={node.get('info')}
              meta={node.get('meta')}
            />
          ))}
        </g>
        <g>
          {labels.map(label => (
            <NodeResourceLabel
              key={label.get('id')}
              label={label.get('label')}
              width={label.get('width')}
              x={label.get('x')}
              y={label.get('y')}
            />
          ))}
        </g>
      </g>
    );
  }
}

function mapStateToProps(state, props) {
  const nodes = layoutNodesByTopologyMetaSelector(state)(state)[props.topologyId];
  // TODO: Move to selectors?
  const labels = getPositionedLabels(nodes, props.transform);
  return { nodes, labels };
}

export default connect(
  mapStateToProps
)(NodesResourcesLayer);
