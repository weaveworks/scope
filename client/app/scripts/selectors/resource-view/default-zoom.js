import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import { RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { canvasMarginsSelector, canvasWidthSelector, canvasHeightSelector } from '../canvas';
import { layersVerticalPositionSelector } from './layers';
import { layoutNodesSelector } from './layout';


// Compute the default zoom settings for the given chart.
export const resourcesDefaultZoomSelector = createSelector(
  [
    layersVerticalPositionSelector,
    layoutNodesSelector,
    canvasMarginsSelector,
    canvasWidthSelector,
    canvasHeightSelector,
  ],
  (layersVerticalPositions, layoutNodes, canvasMargins, width, height) => {
    if (layoutNodes.size === 0) {
      return makeMap();
    }

    const xMin = layoutNodes.map(n => n.get('x')).min();
    const yMin = layersVerticalPositions.min();
    const xMax = layoutNodes.map(n => n.get('x') + n.get('width')).max();
    const yMax = layersVerticalPositions.max() + RESOURCES_LAYER_HEIGHT;

    const minNodeWidth = layoutNodes.map(n => n.get('width')).min();

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
