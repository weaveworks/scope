import React from 'react';
import { connect } from 'react-redux';

import Logo from '../components/logo';
import ZoomContainer from '../components/zoom-container';
import ResourceView from './resource-view';
import NodesChartElements from './nodes-chart-elements';
import { clickBackground } from '../actions/app-actions';
import { isResourceViewModeSelector } from '../selectors/topology';


class NodesChart extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  render() {
    const { isResourceViewMode, isEmpty, selectedNodeId } = this.props;
    const markerOffset = selectedNodeId ? '35' : '40';
    const markerSize = selectedNodeId ? '10' : '30';
    const svgClassNames = isEmpty ? 'hide' : '';

    return (
      <div className="nodes-chart">
        <svg
          width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}
        >
          <defs>
            <marker
              className="edge-marker"
              id="end-arrow"
              viewBox="1 0 10 10"
              refX={markerOffset}
              refY="3.5"
              markerWidth={markerSize}
              markerHeight={markerSize}
              orient="auto"
            >
              <polygon className="link" points="0 0, 10 3.5, 0 7" />
            </marker>
          </defs>
          <g transform="translate(24,24) scale(0.25)">
            <Logo />
          </g>
          <ZoomContainer disabled={selectedNodeId}>
            {isResourceViewMode ? <ResourceView /> : <NodesChartElements />}
          </ZoomContainer>
        </svg>
      </div>
    );
  }

  handleMouseClick() {
    if (this.props.selectedNodeId) {
      this.props.clickBackground();
    }
  }
}


function mapStateToProps(state) {
  return {
    isResourceViewMode: isResourceViewModeSelector(state),
    selectedNodeId: state.get('selectedNodeId'),
  };
}


export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);
