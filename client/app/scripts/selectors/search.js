import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { Map as makeMap } from 'immutable';

import { parseQuery, searchTopology } from '../utils/search-utils';

const nodesByTopologySelector = state => state.get('nodesByTopology');
const currentTopologyIdSelector = state => state.get('currentTopologyId');
const parsedSearchQuerySelector = state => parseQuery(state.get('searchQuery'));

export const searchNodeMatchesSelector = createMapSelector(
  [
    nodesByTopologySelector,
    parsedSearchQuerySelector,
  ],
  (nodes, parsed) => (parsed ? searchTopology(nodes, parsed) : makeMap())
);

export const currentTopologySearchNodeMatchesSelector = createSelector(
  [
    searchNodeMatchesSelector,
    currentTopologyIdSelector,
  ],
  (nodesByTopology, currentTopologyId) => {
    console.log(`>>> Update current topology search nodes matches ${currentTopologyId}`);
    return nodesByTopology.get(currentTopologyId) || makeMap();
  }
);
