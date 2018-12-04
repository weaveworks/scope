import { createSelector } from 'reselect';
import { createMapSelector, createListSelector } from 'reselect-map';
import { fromJS, Map as makeMap, List as makeList } from 'immutable';

import { modulo } from '../utils/math-utils';
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
        .filter(m => !m.get('valueEmpty'))
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

export const availableMetricTypesSelector = createListSelector(
  [
    availableMetricsSelector,
  ],
  metric => metric.get('label')
);

export const pinnedMetricSelector = createSelector(
  [
    availableMetricsSelector,
    state => state.get('pinnedMetricType'),
  ],
  (availableMetrics, metricType) => availableMetrics.find(m => m.get('label') === metricType)
);

export const nextPinnedMetricTypeSelector = createSelector(
  [
    availableMetricTypesSelector,
    state => state.get('pinnedMetricType'),
  ],
  (metricTypes, pinnedMetricType) => {
    const currentIndex = metricTypes.indexOf(pinnedMetricType);
    const nextIndex = modulo(currentIndex + 1, metricTypes.count());
    return metricTypes.get(pinnedMetricType ? nextIndex : 0);
  }
);

export const previousPinnedMetricTypeSelector = createSelector(
  [
    availableMetricTypesSelector,
    state => state.get('pinnedMetricType'),
  ],
  (metricTypes, pinnedMetricType) => {
    const currentIndex = metricTypes.indexOf(pinnedMetricType);
    const previousIndex = modulo(currentIndex - 1, metricTypes.count());
    return metricTypes.get(pinnedMetricType ? previousIndex : 0);
  }
);

export const selectedMetricTypeSelector = createSelector(
  [
    state => state.get('pinnedMetricType'),
    state => state.get('hoveredMetricType'),
  ],
  (pinnedMetricType, hoveredMetricType) => hoveredMetricType || pinnedMetricType
);

const selectedMetricIdSelector = createSelector(
  [
    availableMetricsSelector,
    selectedMetricTypeSelector,
  ],
  (availableMetrics, metricType) =>
    (availableMetrics.find(m => m.get('label') === metricType) || makeMap()).get('id')
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
    selectedMetricIdSelector,
    topCardNodeSelector,
  ],
  (node, selectedMetricId, topCardNode) => {
    const isHighlighted = topCardNode && topCardNode.details && topCardNode.id === node.get('id');
    const sourceNode = isHighlighted ? fromJS(topCardNode.details) : node;
    return sourceNode.get('metrics') && sourceNode.get('metrics')
      .filter(m => m.get('id') === selectedMetricId)
      .first();
  }
);
