import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { enterEdge, leaveEdge } from '../actions/app-actions';
import { encodeIdAttribute, decodeIdAttribute } from '../utils/dom-utils';

function isStorageComponent(id) {
  const storageComponents = ['<persistent_volume>', '<storage_class>', '<persistent_volume_claim>', '<volume_snapshot>', '<volume_snapshot_data>'];
  return storageComponents.includes(id);
}

// getAdjacencyClass takes id which contains information about edge as a topology
// of parent and child node.
// For example: id is of form "nodeA;<storage_class>---nodeB;<persistent_volume_claim>"
function getAdjacencyClass(id) {
  const topologyId = id.split('---');
  const fromNode = topologyId[0].split(';');
  const toNode = topologyId[1].split(';');
  if (fromNode[1] !== undefined && toNode[1] !== undefined) {
    if (isStorageComponent(fromNode[1]) || isStorageComponent(toNode[1])) {
      return 'link-storage';
    }
  }
  return 'link-none';
}

class Edge extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const {
      id, path, highlighted, focused, thickness, source, target
    } = this.props;
    const shouldRenderMarker = (focused || highlighted) && (source !== target);
    const className = classNames('edge', { highlighted });
    return (
      <g
        id={encodeIdAttribute(id)}
        className={className}
        onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave}
      >
        <path className="shadow" d={path} style={{ strokeWidth: 10 * thickness }} />
        <path
          className={getAdjacencyClass(id)}
          d={path}
          style={{ strokeWidth: 5 }}
        />
        <path
          className="link"
          d={path}
          markerEnd={shouldRenderMarker ? 'url(#end-arrow)' : null}
          style={{ strokeWidth: thickness }}
        />
      </g>
    );
  }

  handleMouseEnter(ev) {
    this.props.enterEdge(decodeIdAttribute(ev.currentTarget.id));
  }

  handleMouseLeave(ev) {
    this.props.leaveEdge(decodeIdAttribute(ev.currentTarget.id));
  }
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode')
  };
}

export default connect(
  mapStateToProps,
  { enterEdge, leaveEdge }
)(Edge);
