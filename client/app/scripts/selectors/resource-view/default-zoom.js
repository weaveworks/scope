import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { CANVAS_MARGINS } from '../../constants/styles';
import { viewportWidthSelector, viewportHeightSelector } from '../viewport';
import { layoutNodesSelector } from './layout';


// Compute the default zoom settings for the given chart.
export const resourcesDefaultZoomSelector = createSelector(
  [
    layoutNodesSelector,
    viewportWidthSelector,
    viewportHeightSelector,
  ],
  (layoutNodes, width, height) => {
    if (layoutNodes.size === 0) {
      return makeMap();
    }

    const xMin = layoutNodes.map(n => n.get('x')).min();
    const yMin = layoutNodes.map(n => n.get('y')).min();
    const xMax = layoutNodes.map(n => n.get('x') + n.get('width')).max();
    const yMax = layoutNodes.map(n => n.get('y') + n.get('height')).max();

    const scaleX = (width / (xMax - xMin)) * 0.9;
    const scaleY = (height / (yMax - yMin)) * 0.9;
    const minScale = scaleX * 0.5;
    const maxScale = scaleX * 1000;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scaleX)) / 2) + CANVAS_MARGINS.left;
    const translateY = ((height - ((yMax + yMin) * scaleY)) / 2) + CANVAS_MARGINS.top;

    return makeMap({ scaleX, scaleY, minScale, maxScale, translateX, translateY });
  }
);
