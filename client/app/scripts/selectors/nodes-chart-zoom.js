import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { CANVAS_MARGINS, NODE_BASE_SIZE } from '../constants/styles';
import { activeTopologyZoomCacheKeyPathSelector } from './topology';
import { viewportWidthSelector, viewportHeightSelector } from './canvas-viewport';
import { graphNodesSelector } from './nodes-chart-graph';


// Compute the default zoom settings for the given graph layout.
const defaultZoomSelector = createSelector(
  [
    graphNodesSelector,
    viewportWidthSelector,
    viewportHeightSelector,
  ],
  (graphNodes, width, height) => {
    if (graphNodes.size === 0) {
      return {};
    }

    const xMin = graphNodes.minBy(n => n.get('x')).get('x');
    const xMax = graphNodes.maxBy(n => n.get('x')).get('x');
    const yMin = graphNodes.minBy(n => n.get('y')).get('y');
    const yMax = graphNodes.maxBy(n => n.get('y')).get('y');

    const xFactor = width / (xMax - xMin);
    const yFactor = height / (yMax - yMin);

    // Maximal allowed zoom will always be such that a node covers 1/5 of the viewport.
    const maxZoomScale = Math.min(width, height) / NODE_BASE_SIZE / 5;

    // Initial zoom is such that the graph covers 90% of either the viewport,
    // or one half of maximal zoom constraint, whichever is smaller.
    const zoomScale = Math.min(xFactor, yFactor, maxZoomScale / 2) * 0.9;

    // Finally, we always allow zooming out exactly 5x compared to the initial zoom.
    const minZoomScale = zoomScale / 5;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const panTranslateX = ((width - ((xMax + xMin) * zoomScale)) / 2) + CANVAS_MARGINS.left;
    const panTranslateY = ((height - ((yMax + yMin) * zoomScale)) / 2) + CANVAS_MARGINS.top;

    return { zoomScale, minZoomScale, maxZoomScale, panTranslateX, panTranslateY };
  }
);

const activeLayoutCachedZoomSelector = createSelector(
  [
    state => state.get('zoomCache'),
    activeTopologyZoomCacheKeyPathSelector,
  ],
  (zoomCache, keyPath) => zoomCache.getIn(keyPath.slice(1))
);

// Use the cache to get the last zoom state for the selected topology,
// otherwise use the default zoom options computed from the graph layout.
export const activeLayoutZoomSelector = createSelector(
  [
    activeLayoutCachedZoomSelector,
    defaultZoomSelector,
  ],
  (cachedZoomState, defaultZoomState) => makeMap(cachedZoomState || defaultZoomState)
);
