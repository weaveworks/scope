import { createSelector } from 'reselect';

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
