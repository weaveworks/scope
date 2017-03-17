import React from 'react';
import { connect } from 'react-redux';
import { debounce, pick } from 'lodash';
import { fromJS } from 'immutable';

import { event as d3Event, select } from 'd3-selection';
import { zoom, zoomIdentity } from 'd3-zoom';

import { cacheZoomState } from '../actions/app-actions';
import { transformToString } from '../utils/transform-utils';
import { activeLayoutZoomSelector } from '../selectors/zooming';
import { activeTopologyZoomCacheKeyPathSelector } from '../selectors/topology';
import {
  canvasMarginsSelector,
  canvasWidthSelector,
  canvasHeightSelector,
} from '../selectors/canvas';

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
    this.svg = select(`svg#${this.props.svg}`);

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
    // `forwardTransform` says whether the zoom transform is forwarded to the child
    // component. The advantage of that is more control rendering control in the
    // children, while the disadvantage is that it's slower, as all the children
    // get updated on every zoom/pan action.
    const { children, forwardTransform } = this.props;
    const transform = forwardTransform ? '' : transformToString(this.state);

    return (
      <g className="cachable-zoom-wrapper" transform={transform}>
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

  // Decides which part of the zoom state is cachable depending
  // on the horizontal/vertical degrees of freedom.
  cachableState(state = this.state) {
    // TODO: Probably shouldn't cache the limits if the layout can
    // change a lot. However, before removing them from here, we have
    // to make sure we can always get them from the default zooms.
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

  // Restore the zooming settings
  restoreCachedZoom(props) {
    if (!props.layoutZoom.isEmpty()) {
      const zoomState = props.layoutZoom.toJS();

      // Scaling limits are always set.
      this.zoom = this.zoom.scaleExtent([zoomState.minScale, zoomState.maxScale]);

      // Translation limits are optional.
      if (props.bounded) {
        this.zoom = this.zoom
          // Translation limits are only set if explicitly demanded (currently we are using them
          // in the resource view, but not in the graph view, although I think the idea would be
          // to use them everywhere).
          .translateExtent([
            [zoomState.minTranslateX, zoomState.minTranslateY],
            [zoomState.maxTranslateX, zoomState.maxTranslateY],
          ])
          // This is to ensure that the translation limits are properly
          // centered, so that the canvas margins are respected.
          .extent([
            [props.canvasMargins.left, props.canvasMargins.top],
            [props.canvasMargins.left + props.width, props.canvasMargins.top + props.height]
          ]);
      }

      // After the limits have been set, update the zoom.
      this.svg.call(this.zoom.transform, zoomIdentity
        .translate(zoomState.translateX, zoomState.translateY)
        .scale(zoomState.scaleX, zoomState.scaleY));

      // Update the state variables.
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
    width: canvasWidthSelector(state),
    height: canvasHeightSelector(state),
    canvasMargins: canvasMarginsSelector(state),
    layoutZoom: activeLayoutZoomSelector(state),
    layoutId: JSON.stringify(activeTopologyZoomCacheKeyPathSelector(state)),
    forceRelayout: state.get('forceRelayout'),
  };
}


export default connect(
  mapStateToProps,
  { cacheZoomState }
)(CachableZoomWrapper);
