import { createSelector } from 'reselect';
import { Map as makeMap } from 'immutable';

import {
  RESOURCE_VIEW_MODE,
  GRAPH_VIEW_MODE,
  TABLE_VIEW_MODE,
} from '../constants/naming';


// TODO: Consider moving more stuff from 'topology-utils' here.

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

// This is used by the resource view where we're always taking the nodes from the cache,
// so that polling doesn't affect the layout. Once we implement a more robust polling
// mechanism that could poll multiple topologies at once, we'll be able to get rid of this.
export const cachedCurrentTopologyNodesSelector = createSelector(
  [
    state => state.get('nodesByTopology'),
    state => state.get('currentTopologyId'),
  ],
  (nodesByTopology, currentTopologyId) => nodesByTopology.get(currentTopologyId, makeMap())
);

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
    isGraphViewModeSelector,
    state => state.get('topologyViewMode'),
    state => state.get('currentTopologyId'),
    state => state.get('pinnedMetricType'),
    state => JSON.stringify(activeTopologyOptionsSelector(state)),
  ],
  (isGraphViewMode, viewMode, topologyId, pinnedMetricType, topologyOptions) => (
    isGraphViewMode ?
      // In graph view, selecting different options/filters produces a different layout.
      ['zoomCache', viewMode, topologyId, topologyOptions] :
      // Otherwise we're in the resource view where the options are hidden (for now),
      // but pinning different metrics can result in very different layouts.
      ['zoomCache', viewMode, topologyId, pinnedMetricType]
  )
);
