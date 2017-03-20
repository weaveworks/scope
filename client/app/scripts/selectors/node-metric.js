import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { fromJS, Map as makeMap, List as makeList } from 'immutable';

import {
  isResourceViewModeSelector,
  cachedCurrentTopologyNodesSelector,
} from '../selectors/topology';


// Resource view uses the metrics of the nodes from the cache, while the graph and table
// view are looking at the current nodes (which are among other things filtered by topology
// options which are currently ignored in the resource view).
export const availableMetricsSelector = createSelector(
  [
    isResourceViewModeSelector,
    cachedCurrentTopologyNodesSelector,
    state => state.get('nodes'),
  ],
  (isResourceView, cachedCurrentTopologyNodes, freshNodes) => (
    (isResourceView ? cachedCurrentTopologyNodes : freshNodes)
      .valueSeq()
      .flatMap(n => n.get('metrics', makeList()))
      .map(m => makeMap({ id: m.get('id'), label: m.get('label') }))
      .toSet()
      .toList()
      .sortBy(m => m.get('label'))
  )
);

export const pinnedMetricSelector = createSelector(
  [
    availableMetricsSelector,
    state => state.get('pinnedMetricType'),
  ],
  (availableMetrics, pinnedMetricType) => {
    const metric = availableMetrics.find(m => m.get('label') === pinnedMetricType);
    return metric && metric.get('id');
  }
);

const topCardNodeSelector = createSelector(
  [
    state => state.get('nodeDetails')
  ],
  nodeDetails => nodeDetails.last()
);

export const nodeMetricSelector = createMapSelector(
  [
    state => state.get('nodes'),
    state => state.get('selectedMetric'),
    topCardNodeSelector,
  ],
  (node, selectedMetric, topCardNode) => {
    const isHighlighted = topCardNode && topCardNode.details && topCardNode.id === node.get('id');
    const sourceNode = isHighlighted ? fromJS(topCardNode.details) : node;
    return sourceNode.get('metrics') && sourceNode.get('metrics')
      .filter(m => m.get('id') === selectedMetric)
      .first();
  }
);
