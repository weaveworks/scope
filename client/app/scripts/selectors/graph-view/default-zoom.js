import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { CANVAS_MARGINS, NODE_BASE_SIZE } from '../../constants/styles';
import { viewportWidthSelector, viewportHeightSelector } from '../viewport';
import { graphNodesSelector } from './graph';


// Compute the default zoom settings for the given graph.
export const graphDefaultZoomSelector = createSelector(
  [
    graphNodesSelector,
    viewportWidthSelector,
    viewportHeightSelector,
  ],
  (graphNodes, width, height) => {
    if (graphNodes.size === 0) {
      return makeMap();
    }

    const xMin = graphNodes.map(n => n.get('x') - NODE_BASE_SIZE).min();
    const yMin = graphNodes.map(n => n.get('y') - NODE_BASE_SIZE).min();
    const xMax = graphNodes.map(n => n.get('x') + NODE_BASE_SIZE).max();
    const yMax = graphNodes.map(n => n.get('y') + NODE_BASE_SIZE).max();

    const xFactor = width / (xMax - xMin);
    const yFactor = height / (yMax - yMin);

    // Maximal allowed zoom will always be such that a node covers 1/5 of the viewport.
    const maxScale = Math.min(width, height) / NODE_BASE_SIZE / 5;

    // Initial zoom is such that the graph covers 90% of either the viewport,
    // or one half of maximal zoom constraint, whichever is smaller.
    const scale = Math.min(xFactor, yFactor, maxScale / 2) * 0.9;

    // Finally, we always allow zooming out exactly 5x compared to the initial zoom.
    const minScale = scale / 5;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scale)) / 2) + CANVAS_MARGINS.left;
    const translateY = ((height - ((yMax + yMin) * scale)) / 2) + CANVAS_MARGINS.top;

    return makeMap({ scaleX: scale, scaleY: scale, minScale, maxScale, translateX, translateY });
  }
);
