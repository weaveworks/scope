import { Map as makeMap } from 'immutable';

import { getNodeColor } from '../utils/color-utils';
import { RESOURCES_LAYER_HEIGHT } from '../constants/styles';


export function nodeResourceViewColorDecorator(node) {
  // Color lightness is normally determined from the node label. However, in the resource view
  // mode, we don't want to vary the lightness so we just always forward the empty string instead.
  return node.set('color', getNodeColor(node.get('rank'), '', node.get('pseudo')));
}

// Decorates the resource node with dimensions taken from its metric summary.
export function nodeResourceBoxDecorator(node) {
  const metricSummary = node.get('metricSummary', makeMap());
  const width = metricSummary.get('showCapacity') ?
    metricSummary.get('totalCapacity') :
    metricSummary.get('absoluteConsumption');
  const height = RESOURCES_LAYER_HEIGHT;

  return node.merge(makeMap({ width, height }));
}

// Decorates the node with the summary info of its metric of a fixed type.
export function nodeMetricSummaryDecoratorByType(metricType, showCapacity) {
  return (node) => {
    const metric = node
      .get('metrics', makeMap())
      .find(m => m.get('label') === metricType);

    // Do nothing if there is no metric info.
    if (!metric) return node;

    const absoluteConsumption = metric.get('value');
    const totalCapacity = showCapacity ? metric.get('max') : absoluteConsumption;
    const relativeConsumption = absoluteConsumption / totalCapacity;
    const format = metric.get('format');

    return node.set('metricSummary', makeMap({
      showCapacity, totalCapacity, absoluteConsumption, relativeConsumption, format
    }));
  };
}

// Decorates the node with the ID of the parent node belonging to a fixed topology.
export function nodeParentDecoratorByTopologyId(topologyId) {
  return (node) => {
    const parent = node
      .get('parents', makeMap())
      .find(p => p.get('topologyId') === topologyId);

    return parent ? node.set('parentNodeId', parent.get('id')) : node;
  };
}
