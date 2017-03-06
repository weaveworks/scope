import React from 'react';
import { connect } from 'react-redux';

import { layoutNodesSelector } from '../selectors/resource-view/layout';
import NodeResourceMetric from './node-resource-metric';


class ResourceView extends React.Component {
  render() {
    return (
      <g className="resource-view">
        {this.props.layoutNodes.map(node => (
          <NodeResourceMetric
            key={node.get('id')}
            label={node.get('label')}
            color={node.get('color')}
            width={node.get('width')}
            height={node.get('height')}
            consumption={node.get('consumption')}
            x={node.get('x')}
            y={node.get('y')}
            withCapacity={node.get('withCapacity')}
            meta={node.get('meta')}
          />
        ))}
      </g>
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
)(ResourceView);
