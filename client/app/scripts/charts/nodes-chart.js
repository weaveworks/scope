import React from 'react';
import { connect } from 'react-redux';

import Logo from '../components/logo';
import NodesChartElements from './nodes-chart-elements';
import CachableZoomWrapper from '../components/cachable-zoom-wrapper';
import { clickBackground } from '../actions/app-actions';


class NodesChart extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.handleMouseClick = this.handleMouseClick.bind(this);
  }

  render() {
    // TODO: What to do with empty?
    const { isGraphViewMode, isEmpty, selectedNodeId } = this.props;
    const markerOffset = selectedNodeId ? '35' : '40';
    const markerSize = selectedNodeId ? '10' : '30';
    const svgClassNames = isEmpty ? 'hide' : '';

    return (
      <div className="nodes-chart">
        <svg
          width="100%" height="100%" id="nodes-chart-canvas"
          className={svgClassNames} onClick={this.handleMouseClick}>
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
          <Logo transform="translate(24,24) scale(0.25)" />
          <CachableZoomWrapper fixVertical={!isGraphViewMode} disabled={selectedNodeId}>
            {isGraphViewMode ? <NodesChartElements /> : <ResourceView />}
          </CachableZoomWrapper>
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
    selectedNodeId: state.get('selectedNodeId'),
  };
}


export default connect(
  mapStateToProps,
  { clickBackground }
)(NodesChart);
