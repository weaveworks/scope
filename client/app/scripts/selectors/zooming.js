import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import {
  graphZoomLimitsSelector,
  graphDefaultZoomSelector,
} from './graph-view/default-zoom';
import {
  resourcesZoomLimitsSelector,
  resourcesDefaultZoomSelector,
} from './resource-view/default-zoom';
import {
  activeTopologyZoomCacheKeyPathSelector,
  isGraphViewModeSelector,
} from './topology';


const activeLayoutCachedZoomSelector = createSelector(
  [
    state => state.get('zoomCache'),
    activeTopologyZoomCacheKeyPathSelector,
  ],
  (zoomCache, keyPath) => zoomCache.getIn(keyPath.slice(1), makeMap())
);

export const activeLayoutZoomLimitsSelector = createSelector(
  [
    isGraphViewModeSelector,
    graphZoomLimitsSelector,
    resourcesZoomLimitsSelector,
  ],
  (isGraphView, graphZoomLimits, resourcesZoomLimits) => (
    isGraphView ? graphZoomLimits : resourcesZoomLimits
  )
);

export const activeLayoutZoomStateSelector = createSelector(
  [
    isGraphViewModeSelector,
    graphDefaultZoomSelector,
    resourcesDefaultZoomSelector,
    activeLayoutCachedZoomSelector,
  ],
  (isGraphView, graphDefaultZoom, resourcesDefaultZoom, cachedZoomState) => {
    const defaultZoom = isGraphView ? graphDefaultZoom : resourcesDefaultZoom;
    // All the cached fields override the calculated default ones.
    return defaultZoom.merge(cachedZoomState);
  }
);
