import { createSelector, createStructuredSelector } from 'reselect';
import { fromJS } from 'immutable';

import { layerNodesSelectorFactory } from './layer-factory';
import { resourceViewLayers } from '../../constants/resources';


export const layersTopologyIdsSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
  ],
  topologyId => fromJS(resourceViewLayers[topologyId] || [])
);

export const layoutNodesByTopologyMetaSelector = createSelector(
  [
    layersTopologyIdsSelector,
  ],
  (layersTopologyIds) => {
    const layerSelectorsMap = {};
    let prevSelector = () => null;

    layersTopologyIds.forEach((topId) => {
      layerSelectorsMap[topId] = layerNodesSelectorFactory(topId, prevSelector);
      prevSelector = layerSelectorsMap[topId];
    });

    return createStructuredSelector(layerSelectorsMap);
  }
);

export const layoutNodesSelector = createSelector(
  [
    state => layoutNodesByTopologyMetaSelector(state)(state),
    layersTopologyIdsSelector,
  ],
  (layoutNodesByTopology, layersTopologyIds) => (
    layersTopologyIds.flatMap(topId => layoutNodesByTopology[topId].toList())
  )
);
