import { createSelector, createStructuredSelector } from 'reselect';
import { fromJS, Map as makeMap } from 'immutable';

import { layersVerticalPositionSelector } from './layers';
import { layersDefs, RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { getNodeColor } from '../../utils/color-utils';
/* eslint no-unused-vars: 0 */
/* eslint no-nested-ternary: 0 */
/* eslint no-sequences: 0 */

const basePseudoId = 'base';

// TODO: Make this variable
const getMetric = (node, metricName) => (
  node.get('metrics', makeMap()).find(m => m.get('label') === metricName)
);

export const layerNodesSelectorFactory = (topologyId, parentLayerNodesSelector) => (
  createSelector(
    [
      state => state.getIn(['nodesByTopology', topologyId], makeMap()),
      state => state.get('pinnedMetricType', 'CPU'),
      layersVerticalPositionSelector,
      parentLayerNodesSelector,
    ],
    (nodes, pinnedMetricType, layersVerticalPosition, parentLayerNodes) => {
      const childrenXOffset = { [basePseudoId]: 0 };
      const layerDef = layersDefs[topologyId];
      let positionedNodes = makeMap();

      parentLayerNodes = parentLayerNodes || makeMap({ basePseudoId: makeMap({ x: 0 }) });

      nodes.forEach((node) => {
        const metric = getMetric(node, pinnedMetricType);
        if (!metric) return;

        const nodeId = node.get('id');
        const nodeColor = getNodeColor(node.get('rank'), node.get('label'), node.get('pseudo'));

        const totalCapacity = metric.get('max') / 1e5;
        const absoluteConsumption = metric.get('value') / 1e5
          / (topologyId === 'processes' ? 4 : 1);
        const relativeConsumption = absoluteConsumption / totalCapacity;
        const nodeConsumption = layerDef.withCapacity ? relativeConsumption : 1;

        const nodeWidth = layerDef.withCapacity ? totalCapacity : absoluteConsumption;

        const parents = node.get('parents') || makeMap();
        const parent = parents.find(p => p.get('topologyId') === layerDef.parentTopologyId);
        const parentId = parent ? parent.get('id') : basePseudoId;

        // NOTE: We don't handle uncontained yet.
        if (parentId === basePseudoId && topologyId !== 'hosts') return;

        childrenXOffset[parentId] = childrenXOffset[parentId]
          || parentLayerNodes.getIn([parentId, 'x'], 0);
        const nodeX = childrenXOffset[parentId];
        const nodeY = layersVerticalPosition.get(topologyId);

        // console.log(nodeX, parentId);
        // TODO: Remove.
        if (nodeX === undefined) return;

        childrenXOffset[parentId] += nodeWidth;

        positionedNodes = positionedNodes.set(nodeId, node.merge(makeMap({
          color: nodeColor,
          x: nodeX,
          y: nodeY,
          width: nodeWidth,
          height: RESOURCES_LAYER_HEIGHT,
          consumption: nodeConsumption,
          withCapacity: layerDef.withCapacity,
          info: `CPU usage: ${absoluteConsumption}%`,
          meta: node,
        })));
      });

      return positionedNodes;
    }
  )
);
