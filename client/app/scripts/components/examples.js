/* eslint no-unused-vars: "off" */
import React from 'react';
import _ from 'lodash';
import NodesChart from '../charts/nodes-chart';
import NodesGrid from '../charts/nodes-grid';
import { deltaAdd, makeNodes } from './debug-toolbar';
import { fromJS, Map as makeMap, Set as makeSet } from 'immutable';


function clog(v) {
  console.log(v);
  return v;
}

function randomGraph(n) {
  return makeMap(makeNodes(n, 'ewq', 4, 'hexagon').map(d => [d.id, fromJS(d)]));
}

function deltaAddSimple(name, adjacency = []) {
  return deltaAdd(name, adjacency, 'circle', false, 1, '');
}


function makeIds(n) {
  return _.range(n).map(i => `n${i}`);
}


function disconnectedGraph(n) {
  return makeMap(makeIds(n)
                  .map((id) => deltaAddSimple(id))
                  .map(d => [d.id, fromJS(d)]));
}


function completeGraph(n) {
  const ids = makeIds(n);
  const allEdges = _.flatMap(ids, i => ids.filter(ii => i !== ii).map(ii => [i, ii]));
  const oneWayEdges = allEdges.filter(edge => _.isEqual(edge, _.sortBy(edge)));
  const adjacencyMap = _(oneWayEdges)
    .groupBy(e => e[0])
    .mapValues(edges => edges.map(e => e[1]))
    .value();
  return makeMap(ids
                  .map((id) => deltaAddSimple(id, adjacencyMap[id] || []))
                  .map(d => [d.id, fromJS(d)]));
}


function completeGraphBi(n) {
  const ids = makeIds(n);
  const adjacency = (id) => ids.filter(_id => _id !== id);
  return makeMap(ids
                  .map((id) => deltaAddSimple(id, adjacency(id)))
                  .map(d => [d.id, fromJS(d)]));
}


function flatTree(n) {
  const ids = makeIds(n + 1);
  const p = ids.pop();
  const adjacency = id => id === p ? ids : [];
  return makeMap(ids.concat([p])
                  .map((id) => deltaAddSimple(id, adjacency(id)))
                  .map(d => [d.id, fromJS(d)]));
}

function proxyGraph(n) {
  const ids = makeIds(n * 2 + 1);
  const p = ids.pop();
  const topIds = _.take(ids, n);
  const bottomIds = _.drop(ids, n);
  const adjacencyMap = Object.assign({
    [p]: bottomIds
  }, _.fromPairs(topIds.map(id => [id, [p]])));

  return makeMap(ids.concat([p])
                  .map((id) => deltaAddSimple(id, adjacencyMap[id] || []))
                  .map(d => [d.id, fromJS(d)]));
}


function chart(nodes,
               n,
               style = { width: 250, height: 250 },
               margins = { top: 0, left: 0, right: 0, bottom: 0 },
               nodeSize = null) {
  return (
    <div key={n} className="example-chart" style={style}>
      <NodesChart
        nodes={nodes}
        width={style.width}
        height={style.height}
        margins={margins}
        highlightedNodeIds={makeSet()}
        highlightedEdgeIds={makeSet()}
        layoutPrecision="3"
        topologyId={Math.random()}
        noZoom="true"
        nodeSize={nodeSize}
      />
    </div>
  );
}


function variants() {
  const nCharts = 5;
  const width = 250;
  const style = {width: nCharts * (width + 16)};
  const generators = [
    { id: 'disconnectedGraph', fn: disconnectedGraph },
    { id: 'completeGraphBi', fn: completeGraphBi },
    { id: 'completeGraph', fn: completeGraph },
    { id: 'flatTree', fn: flatTree },
    { id: 'proxyGraph', fn: proxyGraph }
  ];
  return (
    <div>
      {_.reverse(generators).map(({id, fn}) => (
        <div key={id} className="nodes-chart-examples" style={style}>
          {_.range(1, nCharts + 1).map(i => (
            chart(fn(i), i)
          ))}
        </div>
      ))}
    </div>
  );
}


function gridView() {
  const nodes = randomGraph(50).map(node => node.remove('label').remove('label_minor'));
  const nodeSize = 24;
  return (
    <NodesGrid width="500" height="500" nodeSize={nodeSize} nodes={nodes} />
  );
}


export class Examples extends React.Component {
  render() {
    return (
      <div className="examples">
        {gridView()}
        {false && variants()}
      </div>
    );
  }
}
