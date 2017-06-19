import { createSelector } from 'reselect';
import { fromJS } from 'immutable';

import { isResourceViewModeSelector } from './topology';
import { layoutNodesByTopologyIdSelector } from './resource-view/layout';
import { RESOURCE_VIEW_LAYERS } from '../constants/resources';


export const shownNodesSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  nodes => nodes.filter(node => !node.get('filtered'))
);

export const shownResourceTopologyIdsSelector = createSelector(
  [
    layoutNodesByTopologyIdSelector,
  ],
  layoutNodesByTopologyId => layoutNodesByTopologyId.filter(nodes => !nodes.isEmpty()).keySeq()
);

// TODO: Get rid of this logic by unifying `nodes` and `nodesByTopology` global states.
const loadedAllResourceLayersSelector = createSelector(
  [
    state => state.get('nodesByTopology').keySeq(),
  ],
  resourceViewLoadedTopologyIds => fromJS(RESOURCE_VIEW_LAYERS).keySeq()
    .every(topId => resourceViewLoadedTopologyIds.contains(topId))
);

export const nodesLoadedSelector = createSelector(
  [
    state => state.get('nodesLoaded'),
    loadedAllResourceLayersSelector,
    isResourceViewModeSelector,
  ],
  (nodesLoaded, loadedAllResourceLayers, isResourceViewMode) => (
    // Since `nodesLoaded` is set when we receive nodes delta over websockets,
    // it's a completely wrong criterion for determining if resource view is
    // in the loading state - instead we look at the 'static' topologies whose
    // nodes were loaded into 'nodesByTopology' and say resource view has been
    // loaded if nodes for all the resource layer topologies have been loaded once.
    isResourceViewMode ? loadedAllResourceLayers : nodesLoaded
  )
);
