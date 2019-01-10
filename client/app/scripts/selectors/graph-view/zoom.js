import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { NODE_BASE_SIZE } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { activeLayoutCachedZoomSelector } from '../zooming';
import { graphNodesSelector } from './graph';

// Nodes in the layout are always kept between 3px and 200px big.
const MAX_SCALE = 200 / NODE_BASE_SIZE;
const MIN_SCALE = 3 / NODE_BASE_SIZE;

const graphBoundingRectangleSelector = createSelector(
  [
    graphNodesSelector,
  ],
  (graphNodes) => {
    if (graphNodes.size === 0) return null;

    const xMin = graphNodes.map(n => n.get('x') - NODE_BASE_SIZE).min();
    const yMin = graphNodes.map(n => n.get('y') - NODE_BASE_SIZE).min();
    const xMax = graphNodes.map(n => n.get('x') + NODE_BASE_SIZE).max();
    const yMax = graphNodes.map(n => n.get('y') + NODE_BASE_SIZE).max();

    return makeMap({
      xMax, xMin, yMax, yMin
    });
  }
);

// Compute the default zoom settings for the given graph.
export const graphDefaultZoomSelector = createSelector(
  [
    graphBoundingRectangleSelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (boundingRectangle, canvasMargins, width, height) => {
    if (!boundingRectangle) return makeMap();

    const {
      xMin, xMax, yMin, yMax
    } = boundingRectangle.toJS();
    const xFactor = width / (xMax - xMin);
    const yFactor = height / (yMax - yMin);

    // Initial zoom is such that the graph covers 90% of either the viewport,
    // or one half of maximal zoom constraint, whichever is smaller.
    const scale = Math.min(xFactor, yFactor, MAX_SCALE / 2) * 0.9;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scale)) / 2) + canvasMargins.left;
    const translateY = ((height - ((yMax + yMin) * scale)) / 2) + canvasMargins.top;

    return makeMap({
      scaleX: scale,
      scaleY: scale,
      translateX,
      translateY,
    });
  }
);

export const graphLimitsSelector = createSelector(
  [
    graphBoundingRectangleSelector,
  ],
  (boundingRectangle) => {
    if (!boundingRectangle) return makeMap();

    const {
      xMin, xMax, yMin, yMax
    } = boundingRectangle.toJS();

    return makeMap({
      contentMaxX: xMax,
      contentMaxY: yMax,
      contentMinX: xMin,
      contentMinY: yMin,
      maxScale: MAX_SCALE,
      minScale: MIN_SCALE,
    });
  }
);

export const graphZoomStateSelector = createSelector(
  [
    graphDefaultZoomSelector,
    activeLayoutCachedZoomSelector,
  ],
  // All the cached fields override the calculated default ones.
  (graphDefaultZoom, cachedZoomState) => graphDefaultZoom.merge(cachedZoomState)
);
