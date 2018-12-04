import React from 'react';
import { connect } from 'react-redux';

import NodesChartElements from './nodes-chart-elements';
import ZoomableCanvas from '../components/zoomable-canvas';
import { transformToString } from '../utils/transform-utils';
import { clickBackground } from '../actions/app-actions';
import {
  graphLimitsSelector,
  graphZoomStateSelector,
} from '../selectors/graph-view/zoom';

import { CONTENT_INCLUDED } from '../constants/naming';


const EdgeMarkerDefinition = ({ selectedNodeId }) => {
  const markerOffset = selectedNodeId ? '35' : '40';
  const markerSize = selectedNodeId ? '10' : '30';
  return (
    <defs>
      <marker
        className="edge-marker"
        id="end-arrow"
        viewBox="1 0 10 10"
        refX={markerOffset}
        refY="3.5"
        markerWidth={markerSize}
        markerHeight={markerSize}
        orient="auto">
        <polygon className="link" points="0 0, 10 3.5, 0 7" />
      </marker>
    </defs>
  );
};

class NodesChart extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  handleMouseClick() {
    if (this.props.selectedNodeId) {
      this.props.clickBackground();
    }
  }

  renderContent(transform) {
    return (
      <g transform={transformToString(transform)}>
        <EdgeMarkerDefinition selectedNodeId={this.props.selectedNodeId} />
        <NodesChartElements />
      </g>
    );
  }

  render() {
    return (
      <div className="nodes-chart">
        <ZoomableCanvas
          onClick={this.handleMouseClick}
          boundContent={CONTENT_INCLUDED}
          limitsSelector={graphLimitsSelector}
          zoomStateSelector={graphZoomStateSelector}
          disabled={this.props.selectedNodeId}>
          {transform => this.renderContent(transform)}
        </ZoomableCanvas>
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    selectedNodeId: state.get('selectedNodeId'),
  };
}


export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);
