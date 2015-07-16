const _ = require('lodash');
const d3 = require('d3');
const webcola = require('webcola');
const debug = require('debug')('scope:nodes-layout');

const MAX_NODES = 100;

const doLayout = function(nodes, edges, width, height, scale) {
  if (_.size(nodes) > MAX_NODES) {
    debug('Too many nodes to lay out.');
    return null;
  }

  const cola = new webcola.Layout()
    .avoidOverlaps(true)
    .size([width, height]);

  const nodeList = _.values(nodes);
  const edgeList = _.values(edges);

  nodeList.forEach(function(v) {
    v.height = scale(2.25);
    v.width = scale(2.25);
  });

  debug('graph layout for node count: ' + _.size(nodes));

  cola
    .convergenceThreshold(1e-3)
    .nodes(nodeList)
    .links(edgeList)
    // .flowLayout('y', 150)
    // .jaccardLinkLengths(20)
    .start(10, 20, 40);

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
