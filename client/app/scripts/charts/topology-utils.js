
export function updateNodeDegrees(nodes, edges) {
  return nodes.map(node => {
    const nodeId = node.get('id');
    const degree = edges.count(edge => {
      return edge.get('source') === nodeId || edge.get('target') === nodeId;
    });
    return node.set('degree', degree);
  });
}
