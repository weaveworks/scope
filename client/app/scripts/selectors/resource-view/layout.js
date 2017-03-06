import { createSelector } from 'reselect';
// import { createMapSelector } from 'reselect-map';
import { fromJS, Map as makeMap } from 'immutable';

import { resourcesLayers } from '../../constants/styles';
import { getNodeColor } from '../../utils/color-utils';


const basePseudoId = 'base';

// TODO: Make this variable
const getCPUMetric = node => (node.get('metrics') || makeMap()).find(m => m.get('label') === 'CPU');

// TODO: Parse this logic into multiple smarter selectors
export const layoutNodesSelector = createSelector(
  [
    state => state.get('nodesByTopology'),
  ],
  (nodesByTopology) => {
    const result = [];
    const childrenXOffset = { [basePseudoId]: 0 };
    let prevTopologyId = null;
    let y = 0;

    resourcesLayers.forEach((layerDef, layerIndex) => {
      y -= layerDef.frameHeight + layerDef.verticalPadding;

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
    });

    return fromJS(result);
  }
);
