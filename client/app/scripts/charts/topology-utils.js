
export function getDegreeForNodeId(topology, nodeId) {
  let degree = 0;
  topology.forEach(node => {
    if (node.get('id') === nodeId) {
      if (node.get('adjacency')) {
        degree += node.get('adjacency').size;
      }
    } else if (node.get('adjacency') && node.get('adjacency').includes(nodeId)) {
      // FIXME this can still count edges double if both directions exist
      degree++;
    }
  });
  return degree;
}
