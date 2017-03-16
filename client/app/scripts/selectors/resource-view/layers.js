import { times } from 'lodash';
import { fromJS, Map as makeMap } from 'immutable';
import { createSelector } from 'reselect';

import { RESOURCES_LAYER_PADDING, RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { resourceViewLayers, topologiesWithCapacity } from '../../constants/resources';
import {
  nodeResourceViewColorDecorator,
  nodeParentNodeDecorator,
  nodeResourceBoxDecorator,
  nodeActiveMetricDecorator,
} from '../../decorators/node';


const RESOURCE_VIEW_MAX_LAYERS = 3;

const nodeWeight = node => (
  node.get('withCapacity') ?
    -node.getIn(['activeMetric', 'relativeConsumption']) :
    -node.get('width')
);

export const layersTopologyIdsSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
  ],
  topologyId => fromJS(resourceViewLayers[topologyId] || [])
);

export const layersVerticalPositionSelector = createSelector(
  [
    layersTopologyIdsSelector,
  ],
  (topologiesIds) => {
    let yPositions = makeMap();
    let currentY = RESOURCES_LAYER_PADDING;

    topologiesIds.forEach((topologyId) => {
      currentY -= RESOURCES_LAYER_HEIGHT + RESOURCES_LAYER_PADDING;
      yPositions = yPositions.set(topologyId, currentY);
    });

    return yPositions;
  }
);

const decoratedNodesByTopologySelector = createSelector(
  [
    layersTopologyIdsSelector,
    state => state.get('pinnedMetricType'),
    ...times(RESOURCE_VIEW_MAX_LAYERS, index => (
      state => state.getIn(['nodesByTopology', layersTopologyIdsSelector(state).get(index)])
    ))
  ],
  (layersTopologyIds, pinnedMetricType, ...topologiesNodes) => {
    let nodesByTopology = makeMap();
    let lastLayerTopologyId = null;

    topologiesNodes.forEach((topologyNodes, index) => {
      const layerTopologyId = layersTopologyIds.get(index);
      const withCapacity = topologiesWithCapacity.includes(layerTopologyId);
      const decoratedTopologyNodes = (topologyNodes || makeMap())
        .map(node => node.set('directParentTopologyId', lastLayerTopologyId))
        .map(node => node.set('topologyId', layerTopologyId))
        .map(node => node.set('activeMetricType', pinnedMetricType))
        .map(node => node.set('withCapacity', withCapacity))
        .map(nodeResourceViewColorDecorator)
        .map(nodeActiveMetricDecorator)
        .map(nodeResourceBoxDecorator)
        .map(nodeParentNodeDecorator);
      const filteredTopologyNodes = decoratedTopologyNodes
        .filter(node => node.get('parentNodeId') || index === 0)
        .filter(node => node.get('width'));

      nodesByTopology = nodesByTopology.set(layerTopologyId, filteredTopologyNodes);
      lastLayerTopologyId = layerTopologyId;
    });

    return nodesByTopology;
  }
);

export const positionedNodesByTopologySelector = createSelector(
  [
    layersTopologyIdsSelector,
    decoratedNodesByTopologySelector,
  ],
  (layersTopologyIds, decoratedNodesByTopology) => {
    let result = makeMap();

    layersTopologyIds.forEach((layerTopologyId, index) => {
      const decoratedNodes = decoratedNodesByTopology.get(layerTopologyId, makeMap());
      const buckets = decoratedNodes.groupBy(node => node.get('parentNodeId'));

      buckets.forEach((bucket, parentNodeId) => {
        const parentTopologyId = layersTopologyIds.get(index - 1);
        let offset = result.getIn([parentTopologyId, parentNodeId, 'offset'], 0);

        bucket.sortBy(nodeWeight).forEach((node, nodeId) => {
          const positionedNode = node.set('offset', offset);
          result = result.setIn([layerTopologyId, nodeId], positionedNode);
          offset += node.get('width');
        });

        // const offset = result.getIn([parentTopologyId, parentNodeId, 'x'], 0);
        // const overhead =
        //   (x - offset) / result.getIn([parentTopologyId, parentNodeId, 'width'], x);
        // if (overhead > 1) {
        //   console.log(overhead);
        //   bucket.forEach((_, nodeId) => {
        //     const node = result.getIn([layerTopologyId, nodeId]);
        //     result = result.mergeIn([layerTopologyId, nodeId], makeMap({
        //       x: ((node.get('x') - offset) / overhead) + offset,
        //       width: node.get('width') / overhead,
        //     }));
        //   });
        // }
      });
    });

    return result;
  }
);
