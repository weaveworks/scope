import React from 'react';
import { connect } from 'react-redux';

import NodeResourceMetric from './node-resource-metric';

class ResourceView extends React.Component {
  render() {
    const { nodes, transform } = this.props;
    return (
      <g className="resource-view" transform={transform}>
        {nodes.toIndexedSeq().map((node, index) => (
          <NodeResourceMetric
            index={index}
            key={node.get('id')}
            metric={node.get('metrics')}
          />
        ))}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    nodes: state.get('nodes'),
  };
}

export default connect(
  mapStateToProps
)(ResourceView);
