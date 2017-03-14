import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { layersVerticalPositionSelector, positionedNodesByTopologySelector } from './layers';


// Compute the default zoom settings for the given chart.
export const resourcesDefaultZoomSelector = createSelector(
  [
    layersVerticalPositionSelector,
    positionedNodesByTopologySelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (verticalPositions, nodes, canvasMargins, width, height) => {
    if (nodes.size === 0) {
      return makeMap();
    }

    const flattenedNodes = nodes.flatten(true);
    const xMin = flattenedNodes.map(n => n.get('x')).min();
    const yMin = verticalPositions.toList().min();
    const xMax = flattenedNodes.map(n => n.get('x') + n.get('width')).max();
    const yMax = verticalPositions.toList().max() + RESOURCES_LAYER_HEIGHT;

    const minNodeWidth = flattenedNodes.map(n => n.get('width')).min();

    const scaleX = (width / (xMax - xMin)) * 1.0;
    const scaleY = (height / (yMax - yMin)) * 0.7;
    const maxScale = width / minNodeWidth;
    const minScale = scaleX;

    // This translation puts the graph in the center of the viewport, respecting the margins.
    const translateX = ((width - ((xMax + xMin) * scaleX)) / 2) + canvasMargins.left;
    const translateY = ((height - ((yMax + yMin) * scaleY)) / 2) + canvasMargins.top;

    return makeMap({
      minTranslateX: xMin,
      maxTranslateX: xMax,
      minTranslateY: yMin,
      maxTranslateY: yMax,
      translateX,
      translateY,
      minScale,
      maxScale,
      scaleX,
      scaleY,
    });
  }
);
