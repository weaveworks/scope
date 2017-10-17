import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { isGraphViewModeSelector, activeTopologyOptionsSelector } from './topology';


export const activeTopologyZoomCacheKeyPathSelector = createSelector(
  [
    isGraphViewModeSelector,
    state => state.get('topologyViewMode'),
    state => state.get('currentTopologyId'),
    state => state.get('pinnedMetricType'),
    state => JSON.stringify(activeTopologyOptionsSelector(state)),
  ],
  (isGraphViewMode, viewMode, topologyId, pinnedMetricType, topologyOptions) => (
    isGraphViewMode
      // In graph view, selecting different options/filters produces a different layout.
      ? ['zoomCache', viewMode, topologyId, topologyOptions]
      // Otherwise we're in the resource view where the options are hidden (for now),
      // but pinning different metrics can result in very different layouts.
      // TODO: Take `topologyId` into account once the resource
      // view layouts start differing between the topologies.
      : ['zoomCache', viewMode, pinnedMetricType]
  )
);

export const activeLayoutCachedZoomSelector = createSelector(
  [
    state => state.get('zoomCache'),
    activeTopologyZoomCacheKeyPathSelector,
  ],
  (zoomCache, keyPath) => zoomCache.getIn(keyPath.slice(1), makeMap())
);
