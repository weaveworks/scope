import { createSelector } from 'reselect';
import { RESOURCE_VIEW_MODE, GRAPH_VIEW_MODE, TABLE_VIEW_MODE } from '../constants/naming';

// TODO: Consider moving more stuff from 'topology-utils' here.

// Checks if graph complexity is high. Used to trigger
// table view on page load and decide on animations.
export const graphExceedsComplexityThreshSelector = createSelector(
  [
    state => state.getIn(['currentTopology', 'stats', 'node_count']) || 0,
    state => state.getIn(['currentTopology', 'stats', 'edge_count']) || 0,
  ],
  (nodeCount, edgeCount) => (nodeCount + (2 * edgeCount)) > 1000
);

// Options for current topology, sub-topologies share options with parent
export const activeTopologyOptionsSelector = createSelector(
  [
    state => state.getIn(['currentTopology', 'parentId']),
    state => state.get('currentTopologyId'),
    state => state.get('topologyOptions'),
  ],
  (parentTopologyId, currentTopologyId, topologyOptions) => (
    topologyOptions.get(parentTopologyId || currentTopologyId)
  )
);

export const activeTopologyZoomCacheKeyPathSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
    activeTopologyOptionsSelector,
  ],
  (topologyId, topologyOptions) => ['zoomCache', topologyId, JSON.stringify(topologyOptions)]
);


export const isGraphViewModeSelector = createSelector(
  [
    state => state.get('topologyViewMode'),
  ],
  viewMode => viewMode === GRAPH_VIEW_MODE
);

export const isTableViewModeSelector = createSelector(
  [
    state => state.get('topologyViewMode'),
  ],
  viewMode => viewMode === TABLE_VIEW_MODE
);

export const isResourceViewModeSelector = createSelector(
  [
    state => state.get('topologyViewMode'),
  ],
  viewMode => viewMode === RESOURCE_VIEW_MODE
);
