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

function searchTopology(nodes, prefix, query) {
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

/**
 * Returns {topologyId: {nodeId: matches}}
 */
export function updateNodeMatches(state) {
  let query = state.get('searchQuery');
  const prefixQuery = query && query.split(PREFIX_DELIMITER);
  const isPrefixQuery = prefixQuery && prefixQuery.length === 2;
  const isValidPrefixQuery = isPrefixQuery && prefixQuery.every(s => s);

  if (query && (isPrefixQuery === isValidPrefixQuery)) {
    const prefix = isValidPrefixQuery ? prefixQuery[0] : null;
    if (isPrefixQuery) {
      query = prefixQuery[1];
    }
    state.get('topologyUrlsById').forEach((url, topologyId) => {
      const topologyNodes = state.getIn(['nodesByTopology', topologyId]);
      if (topologyNodes) {
        const nodeMatches = searchTopology(topologyNodes, prefix, query);
        state = state.setIn(['searchNodeMatches', topologyId], nodeMatches);
      }
    });
  } else {
    state = state.update('searchNodeMatches', snm => snm.clear());
  }

  return state;
}
