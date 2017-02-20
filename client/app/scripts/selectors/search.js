import { createSelector } from 'reselect';
import { createMapSelector } from 'reselect-map';
import { Map as makeMap } from 'immutable';

import { parseQuery, searchNode, searchTopology, getSearchableFields } from '../utils/search-utils';


const parsedSearchQuerySelector = createSelector(
  [
    state => state.get('searchQuery')
  ],
  searchQuery => parseQuery(searchQuery)
);

export const searchNodeMatchesSelector = createMapSelector(
  [
    state => state.get('nodes'),
    parsedSearchQuerySelector,
  ],
  (node, parsed) => (parsed ? searchNode(node, parsed) : makeMap())
);

export const searchMatchCountByTopologySelector = createMapSelector(
  [
    state => state.get('nodesByTopology'),
    parsedSearchQuerySelector,
  ],
  // TODO: Bring map selectors one level deeper here so that `searchTopology` is
  // not executed against all the topology nodes when the nodes delta is small.
  (nodes, parsed) => (parsed ? searchTopology(nodes, parsed).size : 0)
);

export const searchableFieldsSelector = createSelector(
  [
    state => state.get('nodes'),
  ],
  getSearchableFields
);
