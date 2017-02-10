import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { Map as makeMap } from 'immutable';

import { parseQuery, searchTopology, getSearchableFields } from '../utils/search-utils';


const parsedSearchQuerySelector = createSelector(
  [
    state => state.get('searchQuery')
  ],
  searchQuery => parseQuery(searchQuery)
);

export const searchNodeMatchesSelector = createMapSelector(
  [
    state => state.get('nodesByTopology'),
    parsedSearchQuerySelector,
  ],
  // TODO: Bring map selectors one level deeper here so that `searchTopology` is
  // not executed against all the topology nodes when the nodes delta is small.
  (nodes, parsed) => (parsed ? searchTopology(nodes, parsed) : makeMap())
);

export const currentTopologySearchNodeMatchesSelector = createSelector(
  [
    state => state.get('currentTopologyId'),
    searchNodeMatchesSelector,
  ],
  (currentTopologyId, nodesByTopology) => nodesByTopology.get(currentTopologyId) || makeMap()
);

export const searchableFieldsSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  // TODO: Bring this function in the selectors.
  getSearchableFields
);
