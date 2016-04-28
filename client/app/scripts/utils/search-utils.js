import { Map as makeMap } from 'immutable';

const SEARCH_FIELDS = makeMap({
  label: 'label',
  sublabel: 'label_minor'
});

// TODO make this dynamic based on topology
const SEARCH_TABLES = makeMap({
  dockerlabel: 'docker_label_',
  dockerenv: 'docker_env_',
});

const PREFIX_DELIMITER = ':';
const PREFIX_ALL = 'all';
const PREFIX_ALL_SHORT = 'a';
const PREFIX_METADATA = 'metadata';
const PREFIX_METADATA_SHORT = 'm';

function findNodeMatch(nodeMatches, keyPath, text, query, label) {
  const index = text.indexOf(query);
  if (index > -1) {
    nodeMatches = nodeMatches.setIn(keyPath,
      {text, label, matches: [{start: index, length: query.length}]});
  }
  return nodeMatches;
}

function searchTopology(nodes, searchFields, prefix, query) {
  let nodeMatches = makeMap();
  nodes.forEach((node, nodeId) => {
    // top level fields
    searchFields.forEach((field, label) => {
      const keyPath = [nodeId, label];
      nodeMatches = findNodeMatch(nodeMatches, keyPath, node.get(field), query, label);
    });

    // metadata
    if (node.get('metadata') && (prefix === PREFIX_METADATA || prefix === PREFIX_METADATA_SHORT
      || prefix === PREFIX_ALL || prefix === PREFIX_ALL_SHORT)) {
      node.get('metadata').forEach(field => {
        const keyPath = [nodeId, 'metadata', field.get('id')];
        nodeMatches = findNodeMatch(nodeMatches, keyPath, field.get('value'),
          query, field.get('label'));
      });
    }

    // tables (envvars and labels)
    const tables = node.get('tables');
    if (tables) {
      let searchTables;
      if (prefix === PREFIX_ALL || prefix === PREFIX_ALL_SHORT) {
        searchTables = SEARCH_TABLES;
      } else if (prefix) {
        searchTables = SEARCH_TABLES.filter((field, label) => prefix === label);
      }
      if (searchTables && searchTables.size > 0) {
        searchTables.forEach((searchTable) => {
          const table = tables.find(t => t.get('id') === searchTable);
          if (table && table.get('rows')) {
            table.get('rows').forEach(field => {
              const keyPath = [nodeId, 'metadata', field.get('id')];
              nodeMatches = findNodeMatch(nodeMatches, keyPath, field.get('value'),
                query, field.get('label'));
            });
          }
        });
      }
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
    let searchFields = SEARCH_FIELDS;
    if (isPrefixQuery) {
      query = prefixQuery[1];
      searchFields = searchFields.filter((field, label) => label === prefix);
    }
    state.get('topologyUrlsById').forEach((url, topologyId) => {
      const topologyNodes = state.getIn(['nodesByTopology', topologyId]);
      if (topologyNodes) {
        const nodeMatches = searchTopology(topologyNodes, searchFields, prefix, query);
        state = state.setIn(['searchNodeMatches', topologyId], nodeMatches);
      }
    });
  } else {
    state = state.update('searchNodeMatches', snm => snm.clear());
  }

  return state;
}
