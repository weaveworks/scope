import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { Map as makeMap } from 'immutable';

import { parseQuery, searchTopology, getSearchableFields } from '../utils/search-utils';


const allNodesSelector = state => state.get('nodes');
const nodesByTopologySelector = state => state.get('nodesByTopology');
const currentTopologyIdSelector = state => state.get('currentTopologyId');
const searchQuerySelector = state => state.get('searchQuery');

const parsedSearchQuerySelector = createSelector(
  [
    searchQuerySelector
  ],
  searchQuery => parseQuery(searchQuery)
);

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
  (nodesByTopology, currentTopologyId) => nodesByTopology.get(currentTopologyId) || makeMap()
);

export const searchableFieldsSelector = createSelector(
  [
    allNodesSelector,
  ],
  getSearchableFields
);
