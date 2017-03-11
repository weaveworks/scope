import { createSelector } from 'reselect';
import { Map as makeMap, fromJS } from 'immutable';

import { resourceViewLayers } from '../../constants/resources';
import { RESOURCES_LAYER_PADDING, RESOURCES_LAYER_HEIGHT } from '../../constants/styles';

export const layersTopologyIdsSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
  ],
  topologyId => fromJS(resourceViewLayers[topologyId] || [])
);

export const layersVerticalPositionSelector = createSelector(
  [
    layersTopologyIdsSelector,
  ],
  (topologiesIds) => {
    let yPositions = makeMap();
    let currentY = RESOURCES_LAYER_PADDING;

    topologiesIds.forEach((topologyId) => {
      currentY -= RESOURCES_LAYER_HEIGHT + RESOURCES_LAYER_PADDING;
      yPositions = yPositions.set(topologyId, currentY);
    });

    return yPositions;
  }
);
