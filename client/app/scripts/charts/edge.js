import React from 'react';
import { connect } from 'react-redux';
import classNames from 'classnames';

import { enterEdge, leaveEdge } from '../actions/app-actions';
import { encodeIdAttribute, decodeIdAttribute } from '../utils/dom-utils';

class Edge extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleMouseEnter = this.handleMouseEnter.bind(this);
    this.handleMouseLeave = this.handleMouseLeave.bind(this);
  }

  render() {
    const { id, path, highlighted, focused, thickness, source, target } = this.props;
    const shouldRenderMarker = (focused || highlighted) && (source !== target);
    const className = classNames('edge', { highlighted });

    return (
      <g
        id={encodeIdAttribute(id)} className={className}
        onMouseEnter={this.handleMouseEnter}
        onMouseLeave={this.handleMouseLeave}
      >
        <path className="shadow" d={path} style={{ strokeWidth: 10 * thickness }} />
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
