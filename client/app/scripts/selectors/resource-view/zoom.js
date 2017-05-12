import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { activeLayoutCachedZoomSelector } from '../zooming';
import {
  layerVerticalPositionByTopologyIdSelector,
  layoutNodesByTopologyIdSelector,
} from './layout';


// This is used to determine the maximal zoom factor.
const minNodeWidthSelector = createSelector(
  [
    layoutNodesByTopologyIdSelector,
  ],
  layoutNodes => layoutNodes.flatten(true).map(n => n.get('width')).min()
);

const resourceNodesBoundingRectangleSelector = createSelector(
  [
    layerVerticalPositionByTopologyIdSelector,
    layoutNodesByTopologyIdSelector,
  ],
  (verticalPositions, layoutNodes) => {
    if (layoutNodes.size === 0) return null;

    const flattenedNodes = layoutNodes.flatten(true);
    const xMin = flattenedNodes.map(n => n.get('offset')).min();
    const yMin = verticalPositions.toList().min();
    const xMax = flattenedNodes.map(n => n.get('offset') + n.get('width')).max();
    const yMax = verticalPositions.toList().max() + RESOURCES_LAYER_HEIGHT;

    return makeMap({ xMin, xMax, yMin, yMax });
  }
);

// Compute the default zoom settings for given resources.
export const resourcesDefaultZoomSelector = createSelector(
  [
    resourceNodesBoundingRectangleSelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (boundingRectangle, canvasMargins, width, height) => {
    if (!boundingRectangle) return makeMap();

    const { xMin, xMax, yMin, yMax } = boundingRectangle.toJS();

    // The default scale takes all the available horizontal space and 70% of the vertical space.
    const scaleX = (width / (xMax - xMin)) * 1.0;
    const scaleY = (height / (yMax - yMin)) * 0.7;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scaleX)) / 2) + canvasMargins.left;
    const translateY = ((height - ((yMax + yMin) * scaleY)) / 2) + canvasMargins.top;

    return makeMap({
      translateX,
      translateY,
      scaleX,
      scaleY,
    });
  }
);

export const resourcesZoomLimitsSelector = createSelector(
  [
    resourcesDefaultZoomSelector,
    resourceNodesBoundingRectangleSelector,
    minNodeWidthSelector,
    canvasWidthSelector,
  ],
  (defaultZoom, boundingRectangle, minNodeWidth, width) => {
    if (defaultZoom.isEmpty()) return makeMap();

    const { xMin, xMax, yMin, yMax } = boundingRectangle.toJS();

    return makeMap({
      // Maximal zoom is such that the smallest box takes the whole canvas.
      maxScale: width / Math.max(minNodeWidth, 1),
      // Minimal zoom is equivalent to the initial one, where the whole layout matches the canvas.
      minScale: defaultZoom.get('scaleX'),
      minTranslateX: xMin,
      maxTranslateX: xMax,
      minTranslateY: yMin,
      maxTranslateY: yMax,
    });
  }
);

export const resourcesZoomStateSelector = createSelector(
  [
    resourcesDefaultZoomSelector,
    activeLayoutCachedZoomSelector,
  ],
  // All the cached fields override the calculated default ones.
  (resourcesDefaultZoom, cachedZoomState) => resourcesDefaultZoom.merge(cachedZoomState)
);
