import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { graphDefaultZoomSelector } from './graph-view/default-zoom';
// import { resourceDefaultZoomSelector } from './resource-view/default-zoom';
import {
  activeTopologyZoomCacheKeyPathSelector,
  isResourceViewModeSelector,
  isGraphViewModeSelector,
} from './topology';

const activeLayoutCachedZoomSelector = createSelector(
  [
    state => state.get('zoomCache'),
    activeTopologyZoomCacheKeyPathSelector,
  ],
  (zoomCache, keyPath) => zoomCache.getIn(keyPath.slice(1))
);

// Use the cache to get the last zoom state for the selected topology,
// otherwise use the default zoom options computed from the layout.
export const activeLayoutZoomSelector = createSelector(
  [
    activeLayoutCachedZoomSelector,
    isGraphViewModeSelector,
    isResourceViewModeSelector,
    graphDefaultZoomSelector,
  ],
  (cachedZoomState, isGraphView, isResourceView, graphDefaultZoom) => {
    if (cachedZoomState) return makeMap(cachedZoomState);
    // TODO: Modify this.
    if (isResourceView) return graphDefaultZoom;
    if (isGraphView) return graphDefaultZoom;
    return makeMap();
  }
);
