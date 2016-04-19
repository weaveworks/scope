import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';
import { fromJS } from 'immutable';

import NodeContainer from './node-container';

export default class NodesChartNodes extends React.Component {
  render() {
    const {adjacentNodes, highlightedNodeIds,
      layoutNodes, layoutPrecision, nodeScale, onNodeClick, scale,
      selectedMetric, selectedNodeScale, selectedNodeId, topologyId, topCardNode} = this.props;

    const zoomScale = scale;

    // highlighter functions
    const setHighlighted = node => node.set('highlighted',
      highlightedNodeIds.has(node.get('id')) || selectedNodeId === node.get('id'));
    const setFocused = node => node.set('focused', selectedNodeId
      && (selectedNodeId === node.get('id')
      || (adjacentNodes && adjacentNodes.includes(node.get('id')))));
    const setBlurred = node => node.set('blurred', selectedNodeId && !node.get('focused'));

    // make sure blurred nodes are in the background
    const sortNodes = node => {
      if (node.get('blurred')) {
        return 0;
      }
      if (node.get('highlighted')) {
        return 2;
      }
      return 1;
    };

    // TODO: think about pulling this up into the store.
    const metric = node => {
      const isHighlighted = topCardNode && topCardNode.details && topCardNode.id === node.get('id');
      const sourceNode = isHighlighted ? fromJS(topCardNode.details) : node;
      return sourceNode.get('metrics') && sourceNode.get('metrics')
        .filter(m => m.get('id') === selectedMetric)
        .first();
    };

    const nodesToRender = layoutNodes.toIndexedSeq()
      .map(setHighlighted)
      .map(setFocused)
      .map(setBlurred)
      .sortBy(sortNodes);

    return (
      <g className="nodes-chart-nodes">
        {nodesToRender.map(node => <NodeContainer
          blurred={node.get('blurred')}
          focused={node.get('focused')}
          highlighted={node.get('highlighted')}
          topologyId={topologyId}
          shape={node.get('shape')}
          stack={node.get('stack')}
          onClick={onNodeClick}
          key={node.get('id')}
          id={node.get('id')}
          label={node.get('label')}
          pseudo={node.get('pseudo')}
          nodeCount={node.get('nodeCount')}
          subLabel={node.get('subLabel')}
          metric={metric(node)}
          rank={node.get('rank')}
          layoutPrecision={layoutPrecision}
          selectedNodeScale={selectedNodeScale}
          nodeScale={nodeScale}
          zoomScale={zoomScale}
          dx={node.get('x')}
          dy={node.get('y')} />)}
      </g>
    );
  }
}

reactMixin.onClass(NodesChartNodes, PureRenderMixin);
