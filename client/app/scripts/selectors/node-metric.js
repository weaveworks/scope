import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { fromJS } from 'immutable';


export const pinnedMetricSelector = createSelector(
  [
    state => state.get('availableCanvasMetrics'),
    state => state.get('pinnedMetricType'),
  ],
  (availableMetrics, pinnedMetricType) => (
    availableMetrics.find(metric => metric.get('label') === pinnedMetricType)
  )
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
