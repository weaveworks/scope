import React from 'react';
import { connect } from 'react-redux';
import { debounce, pick } from 'lodash';
import { fromJS } from 'immutable';

import { event as d3Event, select } from 'd3-selection';
import { zoom, zoomIdentity } from 'd3-zoom';

import { cacheZoomState } from '../actions/app-actions';
import { activeLayoutZoomSelector } from '../selectors/zooming';
import { activeTopologyZoomCacheKeyPathSelector } from '../selectors/topology';

import { ZOOM_CACHE_DEBOUNCE_INTERVAL } from '../constants/timer';


class CachableZoomWrapper extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      minTranslateX: 0,
      maxTranslateX: 0,
      minTranslateY: 0,
      maxTranslateY: 0,
      translateX: 0,
      translateY: 0,
      minScale: 1,
      maxScale: 1,
      scaleX: 1,
      scaleY: 1,
    };

    this.debouncedCacheZoom = debounce(this.cacheZoom.bind(this), ZOOM_CACHE_DEBOUNCE_INTERVAL);
    this.zoomed = this.zoomed.bind(this);
  }

  componentDidMount() {
    this.zoomRestored = false;
    this.zoom = zoom().on('zoom', this.zoomed);
    // TODO: Make this correct
    this.svg = select('.nodes-chart svg');

    this.setZoomTriggers(!this.props.disabled);
    this.restoreCachedZoom(this.props);
  }

  componentWillUnmount() {
    this.setZoomTriggers(false);
    this.debouncedCacheZoom.cancel();
  }

  componentWillReceiveProps(nextProps) {
    const layoutChanged = nextProps.layoutId !== this.props.layoutId;
    const disabledChanged = nextProps.disabled !== this.props.disabled;

    // If the layout has changed (either active topology or its options) or
    // relayouting has been requested, stop pending zoom caching event and
    // ask for the new zoom settings to be restored again from the cache.
    if (layoutChanged || nextProps.forceRelayout) {
      this.debouncedCacheZoom.cancel();
      this.zoomRestored = false;
    }

    // If the zooming has been enabled/disabled, update its triggers.
    if (disabledChanged) {
      this.setZoomTriggers(!nextProps.disabled);
    }

    if (!this.zoomRestored) {
      this.restoreCachedZoom(nextProps);
    }
  }

  render() {
    const { children, forwardTransform } = this.props;
    const { translateX, translateY, scaleX, scaleY } = this.state;
    const transform = `translate(${translateX},${translateY}) scale(${scaleX},${scaleY})`;

    // Not passing transform into child components by default for perf reasons.
    return (
      <g className="zoom-container" transform={forwardTransform ? '' : transform}>
        {forwardTransform ? children(this.state) : children}
      </g>
    );
  }

  setZoomTriggers(zoomingEnabled) {
    if (zoomingEnabled) {
      this.svg.call(this.zoom);
    } else {
      this.svg.on('.zoom', null);
    }
  }

  cachableState(state = this.state) {
    let cachableFields = [
      'minTranslateX', 'maxTranslateX',
      'minTranslateY', 'maxTranslateY',
      'minScale', 'maxScale'
    ];
    if (!this.props.fixHorizontal) {
      cachableFields = cachableFields.concat(['scaleX', 'translateX']);
    }
    if (!this.props.fixVertical) {
      cachableFields = cachableFields.concat(['scaleY', 'translateY']);
    }
    return pick(state, cachableFields);
  }

  cacheZoom() {
    this.props.cacheZoomState(fromJS(this.cachableState()));
  }

  restoreCachedZoom(props) {
    if (!props.layoutZoom.isEmpty()) {
      const zoomState = props.layoutZoom.toJS();

      // Restore the zooming settings
      this.zoom = this.zoom
        .scaleExtent([zoomState.minScale, zoomState.maxScale])
        .translateExtent([
          [zoomState.minTranslateX, zoomState.minTranslateY],
          [zoomState.maxTranslateX, zoomState.maxTranslateY],
        ]);
      this.svg.call(this.zoom.transform, zoomIdentity
        .translate(zoomState.translateX, zoomState.translateY)
        .scale(zoomState.scaleX, zoomState.scaleY));

      // Update the state variables
      this.setState(zoomState);
      this.zoomRestored = true;
    }
  }

  zoomed() {
    if (!this.props.disabled) {
      const updatedState = this.cachableState({
        scaleX: d3Event.transform.k,
        scaleY: d3Event.transform.k,
        translateX: d3Event.transform.x,
        translateY: d3Event.transform.y,
      });

      this.setState(updatedState);
      this.debouncedCacheZoom();
    }
  }
}


function mapStateToProps(state) {
  return {
    layoutZoom: activeLayoutZoomSelector(state),
    layoutId: JSON.stringify(activeTopologyZoomCacheKeyPathSelector(state)),
    forceRelayout: state.get('forceRelayout'),
  };
}


export default connect(
  mapStateToProps,
  { cacheZoomState }
)(CachableZoomWrapper);
