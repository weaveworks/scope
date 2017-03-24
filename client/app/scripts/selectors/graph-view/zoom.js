import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { NODE_BASE_SIZE } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { activeLayoutCachedZoomSelector } from '../zooming';
import { graphNodesSelector } from './graph';


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

    return makeMap({ xMin, yMin, xMax, yMax });
  }
);

// Max scale limit will always be such that a node covers 1/5 of the viewport.
const maxScaleSelector = createSelector(
  [
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (width, height) => Math.min(width, height) / NODE_BASE_SIZE / 5
);

// Compute the default zoom settings for the given graph.
export const graphDefaultZoomSelector = createSelector(
  [
    graphBoundingRectangleSelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
    maxScaleSelector,
  ],
  (boundingRectangle, canvasMargins, width, height, maxScale) => {
    if (!boundingRectangle) return makeMap();

    const { xMin, xMax, yMin, yMax } = boundingRectangle.toJS();
    const xFactor = width / (xMax - xMin);
    const yFactor = height / (yMax - yMin);

    // Initial zoom is such that the graph covers 90% of either the viewport,
    // or one half of maximal zoom constraint, whichever is smaller.
    const scale = Math.min(xFactor, yFactor, maxScale / 2) * 0.9;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scale)) / 2) + canvasMargins.left;
    const translateY = ((height - ((yMax + yMin) * scale)) / 2) + canvasMargins.top;

    return makeMap({
      translateX,
      translateY,
      scaleX: scale,
      scaleY: scale,
    });
  }
);

export const graphZoomLimitsSelector = createSelector(
  [
    graphDefaultZoomSelector,
    maxScaleSelector,
  ],
  (defaultZoom, maxScale) => {
    if (defaultZoom.isEmpty()) return makeMap();

    // We always allow zooming out exactly 5x compared to the initial zoom.
    const minScale = defaultZoom.get('scaleX') / 5;

    return makeMap({ minScale, maxScale });
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
