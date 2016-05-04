import { Map as makeMap } from 'immutable';

import { slugify } from './string-utils';

// topolevel search fields
const SEARCH_FIELDS = makeMap({
  label: 'label',
  sublabel: 'label_minor'
});

const PREFIX_DELIMITER = ':';

function matchPrefix(label, prefix) {
  if (label && prefix) {
    return (new RegExp(prefix, 'i')).test(slugify(label));
  }
  return false;
}

function findNodeMatch(nodeMatches, keyPath, text, query, prefix, label) {
  if (!prefix || matchPrefix(label, prefix)) {
    const queryRe = new RegExp(query, 'i');
    const matches = text.match(queryRe);
    if (matches) {
      const firstMatch = matches[0];
      const index = text.search(queryRe);
      nodeMatches = nodeMatches.setIn(keyPath,
        {text, label, matches: [{start: index, length: firstMatch.length}]});
    }
  }
  return nodeMatches;
}

export function searchTopology(nodes, { prefix, query }) {
  let nodeMatches = makeMap();
  nodes.forEach((node, nodeId) => {
    // top level fields
    SEARCH_FIELDS.forEach((field, label) => {
      const keyPath = [nodeId, label];
      nodeMatches = findNodeMatch(nodeMatches, keyPath, node.get(field),
        query, prefix, label);
    });

    // metadata
    if (node.get('metadata')) {
      node.get('metadata').forEach(field => {
        const keyPath = [nodeId, 'metadata', field.get('id')];
        nodeMatches = findNodeMatch(nodeMatches, keyPath, field.get('value'),
          query, prefix, field.get('label'));
      });
    }

    // tables (envvars and labels)
    const tables = node.get('tables');
    if (tables) {
      tables.forEach((table) => {
        if (table.get('rows')) {
          table.get('rows').forEach(field => {
            const keyPath = [nodeId, 'metadata', field.get('id')];
            nodeMatches = findNodeMatch(nodeMatches, keyPath, field.get('value'),
              query, prefix, field.get('label'));
          });
        }
      });
    }
  });
  return nodeMatches;
}

export function parseQuery(query) {
  if (query) {
    const prefixQuery = query.split(PREFIX_DELIMITER);
    const isPrefixQuery = prefixQuery && prefixQuery.length === 2;
    const valid = !isPrefixQuery || prefixQuery.every(s => s);
    if (valid) {
      let prefix = null;
      if (isPrefixQuery) {
        prefix = prefixQuery[0];
        query = prefixQuery[1];
      }
      return {
        query,
        prefix
      };
    }
  }
  return null;
}

/**
 * Returns {topologyId: {nodeId: matches}}
 */
export function updateNodeMatches(state) {
  const parsed = parseQuery(state.get('searchQuery'));
  if (parsed) {
    state.get('topologyUrlsById').forEach((url, topologyId) => {
      const topologyNodes = state.getIn(['nodesByTopology', topologyId]);
      if (topologyNodes) {
        const nodeMatches = searchTopology(topologyNodes, parsed);
        state = state.setIn(['searchNodeMatches', topologyId], nodeMatches);
      }
    });
  } else {
    state = state.update('searchNodeMatches', snm => snm.clear());
  }

  return state;
}

/**
 * Set `filtered:true` in state's nodes if a pinned search matches
 */
export function applyPinnedSearches(state) {
  // clear old filter state
  state = state.update('nodes',
    nodes => nodes.map(node => node.set('filtered', false)));

  const pinnedSearches = state.get('pinnedSearches');
  if (pinnedSearches.size > 0) {
    state.get('pinnedSearches').forEach(query => {
      const parsed = parseQuery(query);
      if (parsed) {
        const nodeMatches = searchTopology(state.get('nodes'), parsed);
        const filteredNodes = state.get('nodes')
          .map(node => node.set('filtered',
            node.get('filtered') // matched by previous pinned search
            || nodeMatches.size === 0 // no match, filter all nodes
            || !nodeMatches.has(node.get('id')))); // filter matches
        state = state.set('nodes', filteredNodes);
      }
    });
  }

  return state;
}
