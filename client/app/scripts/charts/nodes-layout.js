const _ = require('lodash');
const d3 = require('d3');
const webcola = require('webcola');
const debug = require('debug')('scope:nodes-layout');

const MAX_NODES = 100;

const getConstraints = function(nodes) {
  const constraints = [];

  // produce node offsets for each rank
  const nodesByRank = _.groupBy(nodes, 'rank');
  _.each(nodesByRank, function(groupedNodes, rank) {
    if (groupedNodes.length > 1 && rank) {
      const offsets = _.map(groupedNodes, function(node) {
        return {
          node: node.index,
          offset: '0'
        };
      });

      constraints.push({
        type: 'alignment',
        axis: 'y',
        offsets: offsets
      });
    }
  });

  return constraints;
};

const doLayout = function(nodes, edges, width, height, scale) {
  if (_.size(nodes) > MAX_NODES) {
    debug('Too many nodes to lay out.');
    return null;
  }

  if (_.size(nodes) === 0) {
    return {height: 0, width: 0};
  }

  const cola = new webcola.Layout()
    .avoidOverlaps(true)
    .size([width, height]);

  const nodeList = _.sortBy(nodes, function(node) {
    return node.id;
  });
  const edgeList = _.values(edges, function(edge) {
    return edge.id;
  });

  nodeList.forEach(function(v, i) {
    v.height = scale(2.25);
    v.width = scale(2.25);
    v.index = i;
  });

  const constraints = getConstraints(nodes);

  debug('graph layout constraints', constraints);

  cola
    .constraints(constraints)
    .convergenceThreshold(1e-3)
    .nodes(nodeList)
    .links(edgeList)
    .start(5, 20, 10);

  debug('graph layout done');

  const extentX = d3.extent(nodeList, function(n) { return n.x; });
  const extentY = d3.extent(nodeList, function(n) { return n.y; });

  // return object with the width and height of layout

  return {
    left: extentX[0],
    height: extentY[1] - extentY[0],
    top: extentY[0],
    width: extentX[1] - extentX[0]
  };
};

module.exports = {
  doLayout: doLayout
};
