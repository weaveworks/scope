import _ from 'lodash';

export function findTopologyById(subTree, topologyId) {
  let foundTopology;

  subTree.forEach(topology => {
    if (_.endsWith(topology.get('url'), topologyId)) {
      foundTopology = topology;
    }
    if (!foundTopology && topology.has('sub_topologies')) {
      foundTopology = findTopologyById(topology.get('sub_topologies'), topologyId);
    }
  });

  return foundTopology;
}

export function updateNodeDegrees(nodes, edges) {
  return nodes.map(node => {
    const nodeId = node.get('id');
    const degree = edges.count(edge => edge.get('source') === nodeId
      || edge.get('target') === nodeId);
    return node.set('degree', degree);
  });
}

/* set topology.id in place on each topology */
export function updateTopologyIds(topologies) {
  return topologies.map(topology => {
    const result = Object.assign({}, topology);
    result.id = topology.url.split('/').pop();
    if (topology.sub_topologies) {
      result.sub_topologies = updateTopologyIds(topology.sub_topologies);
    }
    return result;
  });
}

// adds ID field to topology (based on last part of URL path) and save urls in
// map for easy lookup
export function setTopologyUrlsById(topologyUrlsById, topologies) {
  let urlMap = topologyUrlsById;
  topologies.forEach(topology => {
    urlMap = urlMap.set(topology.id, topology.url);
    if (topology.sub_topologies) {
      topology.sub_topologies.forEach(subTopology => {
        urlMap = urlMap.set(subTopology.id, subTopology.url);
      });
    }
  });
  return urlMap;
}

export function filterHiddenTopologies(topologies) {
  return topologies.filter(t => (!t.hidden_if_empty || t.stats.node_count > 0 ||
                               t.stats.filtered_nodes > 0));
}
