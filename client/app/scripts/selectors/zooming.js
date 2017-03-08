import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { graphDefaultZoomSelector } from './graph-view/default-zoom';
import { resourcesDefaultZoomSelector } from './resource-view/default-zoom';
import { activeTopologyZoomCacheKeyPathSelector, isGraphViewModeSelector } from './topology';


const activeLayoutCachedZoomSelector = createSelector(
  [
    state => state.get('zoomCache'),
    activeTopologyZoomCacheKeyPathSelector,
  ],
  (zoomCache, keyPath) => zoomCache.getIn(keyPath.slice(1), makeMap())
);

export const activeLayoutZoomSelector = createSelector(
  [
    activeLayoutCachedZoomSelector,
    isGraphViewModeSelector,
    graphDefaultZoomSelector,
    resourcesDefaultZoomSelector,
  ],
  (cachedZoomState, isGraphView, graphDefaultZoom, resourcesDefaultZoom) => {
    const defaultZoom = isGraphView ? graphDefaultZoom : resourcesDefaultZoom;
    return defaultZoom.merge(cachedZoomState);
  }
);
