import React from 'react';
import { connect } from 'react-redux';
import { debounce, pick } from 'lodash';
import { fromJS } from 'immutable';

import { event as d3Event, select } from 'd3-selection';
import { zoom, zoomIdentity } from 'd3-zoom';

import Logo from '../components/logo';
import ResourceView from './resource-view';
import NodesChartElements from './nodes-chart-elements';
import { clickBackground, cacheZoomState } from '../actions/app-actions';
import { activeLayoutZoomSelector } from '../selectors/nodes-chart-zoom';
import {
  activeTopologyZoomCacheKeyPathSelector,
  isResourceViewModeSelector
} from '../selectors/topology';

import { ZOOM_CACHE_DEBOUNCE_INTERVAL } from '../constants/timer';


const ZOOM_CACHE_FIELDS = [
  'zoomScale',
  'minZoomScale',
  'maxZoomScale',
  'panTranslateX',
  'panTranslateY',
];

class NodesChart extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      zoomScale: 0,
      minZoomScale: 0,
      maxZoomScale: 0,
      panTranslateX: 0,
      panTranslateY: 0,
    };

    this.debouncedCacheZoom = debounce(this.cacheZoom.bind(this), ZOOM_CACHE_DEBOUNCE_INTERVAL);
    this.handleMouseClick = this.handleMouseClick.bind(this);
    this.zoomed = this.zoomed.bind(this);
  }

  componentDidMount() {
    // distinguish pan/zoom from click
    this.isZooming = false;
    this.zoomRestored = false;
    this.zoom = zoom().on('zoom', this.zoomed);

    this.svg = select('.nodes-chart svg');
    this.svg.call(this.zoom);

    this.restoreCachedZoom(this.props);
  }

  componentWillUnmount() {
    // undoing .call(zoom)
    this.svg
      .on('mousedown.zoom', null)
      .on('onwheel', null)
      .on('onmousewheel', null)
      .on('dblclick.zoom', null)
      .on('touchstart.zoom', null);

    this.debouncedCacheZoom.cancel();
  }

  componentWillReceiveProps(nextProps) {
    const layoutChanged = nextProps.layoutId !== this.props.layoutId;

    // If the layout has changed (either active topology or its options) or
    // relayouting has been requested, stop pending zoom caching event and
    // ask for the new zoom settings to be restored again from the cache.
    if (layoutChanged || nextProps.forceRelayout) {
      this.debouncedCacheZoom.cancel();
      this.zoomRestored = false;
    }

    if (!this.zoomRestored) {
      this.restoreCachedZoom(nextProps);
    }
  }

  render() {
    // Not passing transform into child components for perf reasons.
    const { panTranslateX, panTranslateY, zoomScale } = this.state;
    const transform = `translate(${panTranslateX}, ${panTranslateY}) scale(${zoomScale})`;
    const svgClassNames = this.props.isEmpty ? 'hide' : '';
    const markerOffset = this.props.selectedNodeId ? '35' : '40';
    const markerSize = this.props.selectedNodeId ? '10' : '30';

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
          {this.props.isResourceViewMode ?
            <ResourceView transform={transform} /> :
            <NodesChartElements transform={transform} />
          }
        </svg>
      </div>
    );
  }

  cacheZoom() {
    const zoomState = pick(this.state, ZOOM_CACHE_FIELDS);
    this.props.cacheZoomState(fromJS(zoomState));
  }

  restoreCachedZoom(props) {
    if (!props.layoutZoom.isEmpty()) {
      const zoomState = props.layoutZoom.toJS();

      // Restore the zooming settings
      this.zoom = this.zoom.scaleExtent([zoomState.minZoomScale, zoomState.maxZoomScale]);
      this.svg.call(this.zoom.transform, zoomIdentity
        .translate(zoomState.panTranslateX, zoomState.panTranslateY)
        .scale(zoomState.zoomScale));

      // Update the state variables
      this.setState(zoomState);
      this.zoomRestored = true;
    }
  }

  handleMouseClick() {
    if (!this.isZooming || this.props.selectedNodeId) {
      this.props.clickBackground();
    } else {
      this.isZooming = false;
    }
  }

  zoomed() {
    this.isZooming = true;
    // don't pan while node is selected
    if (!this.props.selectedNodeId) {
      this.setState({
        panTranslateX: d3Event.transform.x,
        panTranslateY: d3Event.transform.y,
        zoomScale: d3Event.transform.k
      });
      this.debouncedCacheZoom();
    }
  }
}


function mapStateToProps(state) {
  return {
    layoutZoom: activeLayoutZoomSelector(state),
    layoutId: JSON.stringify(activeTopologyZoomCacheKeyPathSelector(state)),
    isResourceViewMode: isResourceViewModeSelector(state),
    selectedNodeId: state.get('selectedNodeId'),
    forceRelayout: state.get('forceRelayout'),
  };
}


export default connect(
  mapStateToProps,
  { clickBackground, cacheZoomState }
)(NodesChart);
