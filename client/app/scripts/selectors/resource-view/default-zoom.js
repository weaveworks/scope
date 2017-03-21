import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { layersVerticalPositionSelector, positionedNodesByTopologySelector } from './layers';


const resourcesBoundingRectangleSelector = createSelector(
  [
    layersVerticalPositionSelector,
    positionedNodesByTopologySelector,
  ],
  (verticalPositions, nodes) => {
    if (nodes.size === 0) return null;

    const flattenedNodes = nodes.flatten(true);
    const xMin = flattenedNodes.map(n => n.get('offset')).min();
    const yMin = verticalPositions.toList().min();
    const xMax = flattenedNodes.map(n => n.get('offset') + n.get('width')).max();
    const yMax = verticalPositions.toList().max() + RESOURCES_LAYER_HEIGHT;

    return makeMap({ xMin, xMax, yMin, yMax });
  }
);

// Compute the default zoom settings for the given chart.
export const resourcesDefaultZoomSelector = createSelector(
  [
    resourcesBoundingRectangleSelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (boundingRectangle, canvasMargins, width, height) => {
    if (!boundingRectangle) return makeMap();

    const { xMin, xMax, yMin, yMax } = boundingRectangle.toJS();

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

const minNodeWidthSelector = createSelector(
  [
    positionedNodesByTopologySelector,
  ],
  nodes => nodes.flatten(true).map(n => n.get('width')).min()
);

export const resourcesZoomLimitsSelector = createSelector(
  [
    resourcesDefaultZoomSelector,
    resourcesBoundingRectangleSelector,
    minNodeWidthSelector,
    canvasWidthSelector,
  ],
  (defaultZoom, boundingRectangle, minNodeWidth, width) => {
    if (defaultZoom.isEmpty()) return makeMap();

    const { xMin, xMax, yMin, yMax } = boundingRectangle.toJS();

    return makeMap({
      maxScale: width / minNodeWidth,
      minScale: defaultZoom.get('scaleX'),
      minTranslateX: xMin,
      maxTranslateX: xMax,
      minTranslateY: yMin,
      maxTranslateY: yMax,
    });
  }
);
