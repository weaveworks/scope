import React from 'react';
import { connect } from 'react-redux';
import { Map as makeMap } from 'immutable';

import { getNodeColor } from '../utils/color-utils';
import NodeResourceMetric from './node-resource-metric';

const basePseudoId = 'base';

const layersDefinition = [{
  topologyId: 'hosts',
  horizontalPadding: 15,
  verticalPadding: 5,
  frameHeight: 200,
  withCapacity: true,
}, {
  topologyId: 'containers',
  horizontalPadding: 0.5,
  verticalPadding: 5,
  frameHeight: 150,
  withCapacity: false,
}, {
  topologyId: 'processes',
  horizontalPadding: 0,
  verticalPadding: 5,
  frameHeight: 100,
  withCapacity: false,
}];

const getCPUMetric = node => (node.get('metrics') || makeMap()).find(m => m.get('label') === 'CPU');

const processedNodes = (nodesByTopology) => {
  const result = [];
  const childrenXOffset = { [basePseudoId]: 0 };
  let prevTopologyId = null;
  let y = 0;

  layersDefinition.forEach((layerDef, layerIndex) => {
    const nodes = nodesByTopology.get(layerDef.topologyId);
    if (!nodes) return;

    nodes.forEach((node) => {
      const metric = getCPUMetric(node);
      if (!metric) return;

      const nodeId = node.get('id');
      const nodeColor = getNodeColor(node.get('rank'), node.get('label'), node.get('pseudo'));

      const totalCapacity = metric.get('max');
      const absoluteConsumption = metric.get('value');
      const relativeConsumption = absoluteConsumption / totalCapacity;
      const nodeConsumption = layerDef.withCapacity ? relativeConsumption : 1;

      const nodeWidth = layerDef.withCapacity ? totalCapacity : absoluteConsumption;
      const nodeHeight = layerDef.frameHeight;

      const shiftX = nodeWidth + layerDef.horizontalPadding;
      const parents = node.get('parents') || makeMap();
      const parent = parents.find(p => p.get('topologyId') === prevTopologyId);
      const parentId = parent ? parent.get('id') : basePseudoId;

      const nodeY = y;
      const nodeX = childrenXOffset[parentId];
      // NOTE: We don't handle uncontained yet.
      if (parentId === basePseudoId && layerIndex > 0) {
        return;
      }

      childrenXOffset[parentId] += shiftX;
      childrenXOffset[nodeId] = nodeX;

      result.push(makeMap({
        id: nodeId,
        color: nodeColor,
        x: nodeX,
        y: nodeY,
        width: nodeWidth,
        height: nodeHeight,
        consumption: nodeConsumption,
        withCapacity: layerDef.withCapacity,
        label: node.get('label'),
        meta: node,
      }));
    });

    prevTopologyId = layerDef.topologyId;
    y += layerDef.frameHeight + layerDef.verticalPadding;
  });

  return result;
};

class ResourceView extends React.Component {
  render() {
    const nodesToRender = processedNodes(this.props.nodesByTopology);

    return (
      <g className="resource-view">
        {nodesToRender.map(node => (
          <NodeResourceMetric
            key={node.get('id')}
            label={node.get('label')}
            color={node.get('color')}
            width={node.get('width')}
            height={node.get('height')}
            consumption={node.get('consumption')}
            x={node.get('x')}
            y={node.get('y')}
            withCapacity={node.get('withCapacity')}
            meta={node.get('meta')}
          />
        ))}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    nodesByTopology: state.get('nodesByTopology'),
  };
}

export default connect(
  mapStateToProps
)(ResourceView);
