import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { fromJS, Map as makeMap, List as makeList } from 'immutable';

import { isGraphViewModeSelector, isResourceViewModeSelector } from '../selectors/topology';
import { RESOURCE_VIEW_METRICS } from '../constants/resources';


// Resource view uses the metrics of the nodes from the cache, while the graph and table
// view are looking at the current nodes (which are among other things filtered by topology
// options which are currently ignored in the resource view).
export const availableMetricsSelector = createSelector(
  [
    isGraphViewModeSelector,
    isResourceViewModeSelector,
    state => state.get('nodes'),
  ],
  (isGraphView, isResourceView, nodes) => {
    // In graph view, we always look through the fresh state
    // of topology nodes to get all the available metrics.
    if (isGraphView) {
      return nodes
        .valueSeq()
        .flatMap(n => n.get('metrics', makeList()))
        .map(m => makeMap({ id: m.get('id'), label: m.get('label') }))
        .toSet()
        .toList()
        .sortBy(m => m.get('label'));
    }

    // In resource view, we're displaying only the hardcoded CPU and Memory metrics.
    // TODO: Make this dynamic as well.
    if (isResourceView) {
      return fromJS(RESOURCE_VIEW_METRICS);
    }

    // Don't show any metrics in the table view mode.
    return makeList();
  }
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
