import debug from 'debug';
import { times } from 'lodash';
import { fromJS, Map as makeMap } from 'immutable';
import { createSelector } from 'reselect';

import { RESOURCES_LAYER_PADDING, RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import {
  RESOURCE_VIEW_MAX_LAYERS,
  RESOURCE_VIEW_LAYERS,
  TOPOLOGIES_WITH_CAPACITY,
} from '../../constants/resources';
import {
  nodeParentDecoratorByTopologyId,
  nodeMetricSummaryDecoratorByType,
  nodeResourceViewColorDecorator,
  nodeResourceBoxDecorator,
} from '../../decorators/node';


const log = debug('scope:nodes-layout');

// Used for ordering the resource nodes.
const resourceNodeConsumptionComparator = (node) => {
  const metricSummary = node.get('metricSummary');
  return metricSummary.get('showCapacity') ?
    -metricSummary.get('relativeConsumption') :
    -metricSummary.get('absoluteConsumption');
};

// A list of topologies shown in the resource view of the active topology (bottom to top).
export const layersTopologyIdsSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
  ],
  topologyId => fromJS(RESOURCE_VIEW_LAYERS[topologyId] || [])
);

// Calculates the resource view layer Y-coordinate for every topology in the resource view.
export const layerVerticalPositionByTopologyIdSelector = createSelector(
  [
    layersTopologyIdsSelector,
  ],
  (topologiesIds) => {
    let yPositions = makeMap();
    let yCumulative = RESOURCES_LAYER_PADDING;

    topologiesIds.forEach((topologyId) => {
      yCumulative -= RESOURCES_LAYER_HEIGHT + RESOURCES_LAYER_PADDING;
      yPositions = yPositions.set(topologyId, yCumulative);
    });

    return yPositions;
  }
);

// Decorate and filter all the nodes to be displayed in the current resource view, except
// for the exact node horizontal offsets which are calculated from the data created here.
const decoratedNodesByTopologySelector = createSelector(
  [
    layersTopologyIdsSelector,
    state => state.get('pinnedMetricType'),
    // Generate the dependencies for this selector programmatically (because we want their
    // number to be customizable directly by changing the constant). The dependency functions
    // here depend on another selector, but this seems to work quite fine. For example, if
    // layersTopologyIdsSelector = ['hosts', 'containers'] and RESOURCE_VIEW_MAX_LAYERS = 3,
    // this code will generate:
    //   [
    //     state => state.getIn(['nodesByTopology', 'hosts'])
    //     state => state.getIn(['nodesByTopology', 'containers'])
    //     state => state.getIn(['nodesByTopology', undefined])
    //   ]
    // which will all be captured by `topologiesNodes` and processed correctly (even for undefined).
    ...times(RESOURCE_VIEW_MAX_LAYERS, index => (
      state => state.getIn(['nodesByTopology', layersTopologyIdsSelector(state).get(index)])
    ))
  ],
  (layersTopologyIds, pinnedMetricType, ...topologiesNodes) => {
    let nodesByTopology = makeMap();
    let parentLayerTopologyId = null;

    topologiesNodes.forEach((topologyNodes, index) => {
      const layerTopologyId = layersTopologyIds.get(index);
      const parentTopologyNodes = nodesByTopology.get(parentLayerTopologyId, makeMap());
      const showCapacity = TOPOLOGIES_WITH_CAPACITY.includes(layerTopologyId);
      const isBaseLayer = (index === 0);

      const nodeParentDecorator = nodeParentDecoratorByTopologyId(parentLayerTopologyId);
      const nodeMetricSummaryDecorator =
        nodeMetricSummaryDecoratorByType(pinnedMetricType, showCapacity);

      // Color the node, deduce its anchor point, dimensions and info about its pinned metric.
      const decoratedTopologyNodes = (topologyNodes || makeMap())
        .map(nodeResourceViewColorDecorator)
        .map(nodeMetricSummaryDecorator)
        .map(nodeResourceBoxDecorator)
        .map(nodeParentDecorator);

      const filteredTopologyNodes = decoratedTopologyNodes
        // Filter out the nodes with no parent in the topology of the previous layer, as their
        // positions in the layout could not be determined. The exception is the base layer.
        // TODO: Also make an exception for uncontained nodes (e.g. processes).
        .filter(node => parentTopologyNodes.has(node.get('parentNodeId')) || isBaseLayer)
        // Filter out the nodes with no metric summary data, which is needed to render the node.
        .filter(node => node.get('metricSummary'))
        // Filter out nodes with zero-width, as they will never be shown, no matter how much we
        // zoom in. The example is every node consuming less than 0.005% CPU, as the 2-digit
        // precision will round it down to zero.
        .filter(node => node.get('width') > 0);

      nodesByTopology = nodesByTopology.set(layerTopologyId, filteredTopologyNodes);
      parentLayerTopologyId = layerTopologyId;
    });

    return nodesByTopology;
  }
);

// Calculate (and fix) the offsets for all the displayed resource nodes.
export const layoutNodesByTopologyIdSelector = createSelector(
  [
    layersTopologyIdsSelector,
    decoratedNodesByTopologySelector,
  ],
  (layersTopologyIds, nodesByTopology) => {
    let layoutNodes = makeMap();
    let parentTopologyId = null;

    // Calculate the offsets bottom-to top as each layer needs to know exact offsets of its parents.
    layersTopologyIds.forEach((layerTopologyId) => {
      let positionedNodes = makeMap();

      // Get the nodes in the current layer grouped by their parent nodes.
      // Each of those buckets will be positioned and sorted independently.
      const nodesByParent = nodesByTopology
        .get(layerTopologyId, makeMap())
        .groupBy(n => n.get('parentNodeId'));

      nodesByParent.forEach((nodesBucket, parentNodeId) => {
        // Set the initial offset to the offset of the parent (that has already been set).
        // If there is no offset information, i.e. we're processing the base layer, set it to 0.
        const parentNode = layoutNodes.getIn([parentTopologyId, parentNodeId], makeMap());
        let currentOffset = parentNode.get('offset', 0);

        // Sort the nodes in the current bucket and lay them down one after another.
        nodesBucket.sortBy(resourceNodeConsumptionComparator).forEach((node, nodeId) => {
          const positionedNode = node.set('offset', currentOffset);
          positionedNodes = positionedNodes.set(nodeId, positionedNode);
          currentOffset += node.get('width');
        });

        // TODO: This block of code checks for the overlaps which are caused by children
        // consuming more resources than their parent node. This happens due to inconsistent
        // data being sent from the backend and it needs to be fixed there.
        const parentOffset = parentNode.get('offset', 0);
        const parentWidth = parentNode.get('width', currentOffset);
        const totalChildrenWidth = currentOffset - parentOffset;
        // If the total width of the children exceeds the parent node box width, we have a problem.
        // We fix it by shrinking all the children to by a factor to perfectly fit into the parent.
        if (totalChildrenWidth > parentWidth) {
          const shrinkFactor = parentWidth / totalChildrenWidth;
          log(`Inconsistent data: Children of ${parentNodeId} reported to use more ` +
            `resource than the node itself - shrinking by factor ${shrinkFactor}`);
          // Shrink all the children.
          nodesBucket.forEach((_, nodeId) => {
            let node = positionedNodes.get(nodeId);
            // Shrink the width of the resource box and update its relative offset.
            node = node.merge(makeMap({
              offset: ((node.get('offset') - parentOffset) * shrinkFactor) + parentOffset,
              width: node.get('width') * shrinkFactor,
            }));
            // Update the metrics summary to reflect the adjusted dimensions for consistent data.
            node = nodeMetricSummaryDecoratorByType(
              node.getIn(['metricSummary', 'type']),
              node.get('showCapacity'),
              shrinkFactor
            )(node);
            // Update the node in the layout.
            positionedNodes = positionedNodes.mergeIn([nodeId], node);
          });
        }
      });

      // Update the layout with the positioned node from the current layer.
      layoutNodes = layoutNodes.mergeIn([layerTopologyId], positionedNodes);
      parentTopologyId = layerTopologyId;
    });

    return layoutNodes;
  }
);
