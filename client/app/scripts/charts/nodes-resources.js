import React from 'react';
import { connect } from 'react-redux';

import Logo from '../components/logo';
import { layoutNodesSelector } from '../selectors/resource-view/layout';
import CachableZoomWrapper from '../components/cachable-zoom-wrapper';
import NodeResourceMetric from './node-resource-metric';


const applyTransform = (node, { scaleX, scaleY, translateX, translateY }) => (
  node.merge({
    x: (node.get('x') * scaleX) + translateX,
    y: (node.get('y') * scaleY) + translateY,
    width: node.get('width') * scaleX,
    height: node.get('height') * scaleY,
  })
);

class NodesResources extends React.Component {
  render() {
    return (
      <div className="nodes-chart">
        <svg
          width="100%" height="100%"
          id="nodes-chart-canvas">
          <Logo transform="translate(24,24) scale(0.25)" />
          <CachableZoomWrapper forwardTransform fixVertical>
            {transform => (
              this.props.layoutNodes.map(node => applyTransform(node, transform)).map(node => (
                <NodeResourceMetric
                  key={node.get('id')}
                  label={node.get('label')}
                  color={node.get('color')}
                  width={node.get('width')}
                  height={node.get('height')}
                  consumption={node.get('consumption')}
                  withCapacity={node.get('withCapacity')}
                  x={node.get('x')}
                  y={node.get('y')}
                  info={node.get('info')}
                  meta={node.get('meta')}
                />
            )))}
          </CachableZoomWrapper>
        </svg>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    layoutNodes: layoutNodesSelector(state),
  };
}

export default connect(
  mapStateToProps
)(NodesResources);
