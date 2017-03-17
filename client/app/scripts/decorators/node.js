import { Map as makeMap } from 'immutable';

import { getNodeColor } from '../utils/color-utils';
import { getMetricValue } from '../utils/metric-utils';
import { RESOURCES_LAYER_HEIGHT } from '../constants/styles';


export function nodeResourceViewColorDecorator(node) {
  // Color lightness is normally determined from the node label. However, in the resource view
  // mode, we don't want to vary the lightness so we just always forward the empty string instead.
  return node.set('color', getNodeColor(node.get('rank'), '', node.get('pseudo')));
}

export function nodeActiveMetricDecorator(node) {
  const metricType = node.get('activeMetricType');
  const metric = node.get('metrics', makeMap()).find(m => m.get('label') === metricType);
  if (!metric) return node;

  const { formattedValue } = getMetricValue(metric);
  const info = `${metricType} - ${formattedValue}`;
  const absoluteConsumption = metric.get('value');
  const withCapacity = node.get('withCapacity');
  const totalCapacity = withCapacity ? metric.get('max') : absoluteConsumption;
  const relativeConsumption = absoluteConsumption / totalCapacity;
  const format = metric.get('format');

  return node.set('activeMetric', makeMap({
    totalCapacity, absoluteConsumption, relativeConsumption, withCapacity, info, format
  }));
}

export function nodeResourceBoxDecorator(node) {
  const widthCriterion = node.get('withCapacity') ? 'totalCapacity' : 'absoluteConsumption';
  const width = node.getIn(['activeMetric', widthCriterion]) * 1e-4;
  const height = RESOURCES_LAYER_HEIGHT;

  return node.merge(makeMap({ width, height }));
}

export function nodeParentNodeDecorator(node) {
  const parentTopologyId = node.get('directParentTopologyId');
  const parents = node.get('parents', makeMap());
  const parent = parents.find(p => p.get('topologyId') === parentTopologyId);
  if (!parent) return node;

  return node.set('parentNodeId', parent.get('id'));
}
