import { createSelector, createStructuredSelector } from 'reselect';

import { layersTopologyIdsSelector } from './layers';
import { layerNodesSelectorFactory } from './layer-factory';


export const layoutNodesByTopologyMetaSelector = createSelector(
  [
    layersTopologyIdsSelector,
  ],
  (layersTopologyIds) => {
    const layerSelectorsMap = {};
    let prevSelector = () => null;
    let prevTopId = null;

    layersTopologyIds.forEach((topId) => {
      layerSelectorsMap[topId] = layerNodesSelectorFactory(topId, prevTopId, prevSelector);
      prevSelector = layerSelectorsMap[topId];
      prevTopId = topId;
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
