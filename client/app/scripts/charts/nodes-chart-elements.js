import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import NodesChartEdges from './nodes-chart-edges';
import NodesChartNodes from './nodes-chart-nodes';

export default class NodesChartElements extends React.Component {
  render() {
    const props = this.props;
    return (
      <g className="nodes-chart-elements" transform={props.transform}>
        <NodesChartEdges layoutEdges={props.layoutEdges}
          layoutPrecision={props.layoutPrecision} />
        <NodesChartNodes layoutNodes={props.layoutNodes} nodeScale={props.nodeScale}
          scale={props.scale} selectedNodeScale={props.selectedNodeScale}
          layoutPrecision={props.layoutPrecision} />
      </g>
    );
  }
}

reactMixin.onClass(NodesChartElements, PureRenderMixin);
