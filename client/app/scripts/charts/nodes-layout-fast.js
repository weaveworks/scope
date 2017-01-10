import { List as makeList, OrderedMap as makeOrderedMap } from 'immutable';

class Bucket {
  constructor(row) {
    this.nodes = {};
    this.row = row;
  }

  size() {
    return Object.keys(this.nodes).length;
  }

  update(node, ref) {
    const nodeId = node.get('id');
    this.nodes[nodeId] = node;
    ref[nodeId] = this;
  }

  remove(node) {
    const nodeId = node.get('id');
    delete this.nodes[nodeId];
  }

  updateLayoutPositions(nodes) {
    const nodeY = (this.row * 50) + 300;
    const shift = Math.ceil(this.size() * 0.5);
    Object.entries(this.nodes).forEach(([key, node], col) => {
      const nodeX = (((col - shift) + ((this.row % 2) * 0.5)) * 50) + 300;
      node = node.merge({ x: nodeX, y: nodeY });
      nodes = nodes.set(node.get('id'), node);
      this.nodes[node.get('id')] = node;
    });
    return nodes;
  }
}

function requiresLayoutReset(addedNodes) {
  return addedNodes.size > 50;
}

class LayoutState {
  constructor() {
    this.cachedNodes = makeOrderedMap();
    this.resetBuckets();
  }

  resetBuckets() {
    this.buckets = [];
    this.bucketCount = 0;
    this.nodeBucketLookup = {};
    this.addEmptyBucket();
  }

  addEmptyBucket() {
    this.buckets.push(new Bucket(this.bucketCount));
    this.bucketCount += 1;
  }

  lastBucket() {
    return this.buckets[this.bucketCount - 1];
  }

  createOptimalLayout(nodes) {
    this.resetBuckets();

    const bucketsLimit = Math.ceil(Math.sqrt(nodes.size));
    nodes.forEach((node) => {
      if (this.lastBucket().size() >= bucketsLimit) {
        this.addEmptyBucket();
      }
      this.lastBucket().update(node, this.nodeBucketLookup);
    });
  }

  nodeBucketHeuristics() {
    return Math.floor(Math.random() * this.buckets.length);
  }

  getNodesDiff(next) {
    const prev = this.cachedNodes;
    const prevIds = prev.map(node => node.get('id'));
    const nextIds = next.map(node => node.get('id'));
    return {
      addedNodes: next.filterNot(node => prevIds.includes(node.get('id'))),
      removedNodes: prev.filterNot(node => nextIds.includes(node.get('id')))
    };
  }

  updateNodes(nodes, forceRelayout) {
    const { addedNodes, removedNodes } = this.getNodesDiff(nodes);

    if (forceRelayout || requiresLayoutReset(addedNodes)) {
      this.createOptimalLayout(nodes);
      this.buckets.forEach((bucket) => {
        nodes = bucket.updateLayoutPositions(nodes);
      });
    } else {
      nodes = this.cachedNodes;
      addedNodes.forEach((node) => {
        const bucket = this.buckets[this.nodeBucketHeuristics()];
        bucket.update(node, this.nodeBucketLookup);
        nodes = bucket.updateLayoutPositions(nodes);
      });
      removedNodes.forEach((node) => {
        const bucket = this.nodeBucketLookup[node.get('id')];
        bucket.remove(node);
        nodes = bucket.updateLayoutPositions(nodes);
      });
    }

    this.cachedNodes = nodes;
    return nodes;
  }

  updateEdges(edges) {
    edges.forEach((edge) => {
      const source = this.cachedNodes.get(edge.get('source'));
      const target = this.cachedNodes.get(edge.get('target'));

      const points = makeList([
        { x: source.get('x'), y: source.get('y') },
        { x: target.get('x'), y: target.get('y') }
      ]);
      edges = edges.setIn([edge.get('id'), 'points'], points);
    });

    return edges;
  }
}

const layoutState = new LayoutState();

export function doLayout(immNodes, immEdges, options) {
  const nodes = layoutState.updateNodes(immNodes, options.forceRelayout);
  const edges = layoutState.updateEdges(immEdges);

  return {
    graphWidth: 1000,
    graphHeight: 1000,
    width: 1000,
    height: 1000,
    nodes,
    edges
  };
}
