import { createSelector, createStructuredSelector } from 'reselect';
import { fromJS, Map as makeMap } from 'immutable';

import { layersDefs } from '../../constants/styles';
import { getNodeColor } from '../../utils/color-utils';
/* eslint no-unused-vars: 0 */
/* eslint no-nested-ternary: 0 */
/* eslint no-sequences: 0 */

const basePseudoId = 'base';

// TODO: Make this variable
const getCPUMetric = node => (node.get('metrics') || makeMap()).find(m => m.get('label') === 'CPU');

export const layerNodesSelectorFactory = (topologyId, parentLayerNodesSelector) => (
  createSelector(
    [
      state => state.getIn(['nodesByTopology', topologyId], makeMap()),
      parentLayerNodesSelector,
    ],
    (nodes, parentLayerNodes) => {
      const childrenXOffset = { [basePseudoId]: 0 };
      const layerDef = layersDefs[topologyId];
      let positionedNodes = makeMap();

      parentLayerNodes = parentLayerNodes || makeMap({ basePseudoId: makeMap({ x: 0 }) });

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

        const parents = node.get('parents') || makeMap();
        const parent = parents.find(p => p.get('topologyId') === layerDef.parentTopologyId);
        const parentId = parent ? parent.get('id') : basePseudoId;

        // NOTE: We don't handle uncontained yet.
        if (parentId === basePseudoId && topologyId !== 'hosts') return;

        childrenXOffset[parentId] = childrenXOffset[parentId]
          || parentLayerNodes.getIn([parentId, 'x'], 0);
        const nodeX = childrenXOffset[parentId];
        const nodeY = topologyId === 'hosts' ? 0 : (topologyId === 'containers' ? -160 : -265);

        // console.log(nodeX, parentId);
        // TODO: Remove.
        if (nodeX === undefined) return;

        childrenXOffset[parentId] += nodeWidth;

        positionedNodes = positionedNodes.set(nodeId, node.merge(makeMap({
          color: nodeColor,
          x: nodeX,
          y: nodeY,
          width: nodeWidth,
          height: nodeHeight,
          consumption: nodeConsumption,
          withCapacity: layerDef.withCapacity,
          meta: node,
        })));
      });

      return positionedNodes;
    }
  )
);
