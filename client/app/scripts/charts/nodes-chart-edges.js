import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import EdgeContainer from './edge-container';

export default class NodesChartEdges extends React.Component {
  render() {
    const {hasSelectedNode, highlightedEdgeIds, layoutEdges, layoutPrecision,
      selectedNodeId} = this.props;

    return (
      <g className="nodes-chart-edges">
        {layoutEdges.toIndexedSeq().map(edge => {
          const sourceSelected = selectedNodeId === edge.get('source');
          const targetSelected = selectedNodeId === edge.get('target');
          const blurred = hasSelectedNode && !sourceSelected && !targetSelected;
          const focused = hasSelectedNode && (sourceSelected || targetSelected);

          return (
            <EdgeContainer
              key={edge.get('id')}
              id={edge.get('id')}
              source={edge.get('source')}
              target={edge.get('target')}
              points={edge.get('points')}
              blurred={blurred}
              focused={focused}
              layoutPrecision={layoutPrecision}
              highlighted={highlightedEdgeIds.has(edge.get('id'))}
            />
          );
        })}
      </g>
    );
  }
}

reactMixin.onClass(NodesChartEdges, PureRenderMixin);
