import { List as makeList } from 'immutable';

export function doLayout(immNodes) {
  let nodes = immNodes;
  const edges = makeList(); // immEdges;

  const columns = Math.ceil(Math.sqrt(nodes.size));
  let row = 0;
  let col = 0;
  let singleX;
  let singleY;
  nodes = nodes.map((node) => {
    if (col === columns) {
      col = 0;
      row += 1;
    }
    singleX = ((col + ((row % 2) * 0.5)) * 50) + 300;
    singleY = (row * 50) + 300;
    col += 1;
    return node.merge({
      x: singleX,
      y: singleY
    });
  });

  // immEdges.forEach((edge) => {
  //   let points = makeList();
  //
  //   // set beginning and end points to node coordinates to ignore node bounding box
  //   const source = nodes.get(edge.get('source'));
  //   const target = nodes.get(edge.get('target'));
  //   points = points.mergeIn([0], {x: source.get('x'), y: source.get('y')});
  //   points = points.mergeIn([1], {x: target.get('x'), y: target.get('y')});
  //
  //   edges = edges.setIn([edge.get('id'), 'points'], points);
  // });

  return {
    graphWidth: 1000,
    graphHeight: 1000,
    width: 1000,
    height: 1000,
    nodes,
    edges
  };
}
