import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { clamp, debounce, pick } from 'lodash';
import { fromJS } from 'immutable';

import { drag } from 'd3-drag';
import { event as d3Event, select } from 'd3-selection';
import { zoomFactor } from 'weaveworks-ui-components/lib/utils/zooming';

import Logo from '../components/logo';
import ZoomControl from '../components/zoom-control';
import { cacheZoomState } from '../actions/app-actions';
import { applyTransform, inverseTransform } from '../utils/transform-utils';
import { activeTopologyZoomCacheKeyPathSelector } from '../selectors/zooming';
import {
  canvasMarginsSelector,
  canvasWidthSelector,
  canvasHeightSelector,
} from '../selectors/canvas';

import { ZOOM_CACHE_DEBOUNCE_INTERVAL } from '../constants/timer';
import { CONTENT_INCLUDED, CONTENT_COVERING } from '../constants/naming';


class ZoomableCanvas extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      contentMaxX: 0,
      contentMaxY: 0,
      contentMinX: 0,
      contentMinY: 0,
      isPanning: false,
      maxScale: 1,
      minScale: 1,
      scaleX: 1,
      scaleY: 1,
      translateX: 0,
      translateY: 0,
    };

    this.debouncedCacheZoom = debounce(this.cacheZoom.bind(this), ZOOM_CACHE_DEBOUNCE_INTERVAL);
    this.handleZoomControlAction = this.handleZoomControlAction.bind(this);
    this.canChangeZoom = this.canChangeZoom.bind(this);

    this.handleZoom = this.handleZoom.bind(this);
    this.handlePanStart = this.handlePanStart.bind(this);
    this.handlePanEnd = this.handlePanEnd.bind(this);
    this.handlePan = this.handlePan.bind(this);
  }

  componentDidMount() {
    this.svg = select('.zoomable-canvas svg');
    this.drag = drag()
      .on('start', this.handlePanStart)
      .on('end', this.handlePanEnd)
      .on('drag', this.handlePan);
    this.svg.call(this.drag);

    this.zoomRestored = false;

    this.updateZoomLimits(this.props);
    this.restoreZoomState(this.props);
    document
      .getElementById('canvas')
      .addEventListener('wheel', this.handleZoom, { passive: false });
  }

  componentWillUnmount() {
    this.debouncedCacheZoom.cancel();
    document
      .getElementById('canvas')
      .removeEventListener('wheel', this.handleZoom, { passive: false });
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

    this.updateZoomLimits(nextProps);
    if (!this.zoomRestored) {
      this.restoreZoomState(nextProps);
    }
  }

  handleZoomControlAction(scale) {
    // Get the center of the SVG and zoom around it.
    const {
      top, bottom, left, right
    } = this.svg.node().getBoundingClientRect();
    const centerOfCanvas = {
      x: (left + right) / 2,
      y: (top + bottom) / 2,
    };
    // Zoom factor diff is obtained by dividing the new zoom scale with the old one.
    this.zoomAtPositionByFactor(centerOfCanvas, scale / this.state.scaleX);
  }

  render() {
    const className = classNames({ panning: this.state.isPanning });

    return (
      <div className="zoomable-canvas">
        <svg id="canvas" className={className} onClick={this.props.onClick}>
          <Logo transform="translate(24,24) scale(0.25)" />
          <g className="zoom-content">
            {this.props.children(this.state)}
          </g>
        </svg>
        {this.canChangeZoom() && <ZoomControl
          zoomAction={this.handleZoomControlAction}
          minScale={this.state.minScale}
          maxScale={this.state.maxScale}
          scale={this.state.scaleX}
        />}
      </div>
    );
  }

  // Decides which part of the zoom state is cachable depending
  // on the horizontal/vertical degrees of freedom.
  cachableState(state = this.state) {
    const cachableFields = []
      .concat(this.props.fixHorizontal ? [] : ['scaleX', 'translateX'])
      .concat(this.props.fixVertical ? [] : ['scaleY', 'translateY']);

    return pick(state, cachableFields);
  }

  cacheZoom() {
    this.props.cacheZoomState(fromJS(this.cachableState()));
  }

  updateZoomLimits(props) {
    this.setState(props.layoutLimits.toJS());
  }

  // Restore the zooming settings
  restoreZoomState(props) {
    if (!props.layoutZoomState.isEmpty()) {
      const zoomState = props.layoutZoomState.toJS();

      // Update the state variables.
      this.setState(zoomState);
      this.zoomRestored = true;
    }
  }

  canChangeZoom() {
    const { disabled, layoutLimits } = this.props;
    const canvasHasContent = !layoutLimits.isEmpty();
    return !disabled && canvasHasContent;
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.setState({ isPanning: false });
  }

  handlePan() {
    let { state } = this;
    // Apply the translation respecting the boundaries.
    state = this.clampedTranslation({
      ...state,
      translateX: this.state.translateX + d3Event.dx,
      translateY: this.state.translateY + d3Event.dy,
    });
    this.updateState(state);
  }

  handleZoom(ev) {
    if (this.canChangeZoom()) {
      // Get the exact mouse cursor position in the SVG and zoom around it.
      const { top, left } = this.svg.node().getBoundingClientRect();
      const mousePosition = {
        x: ev.clientX - left,
        y: ev.clientY - top,
      };
      this.zoomAtPositionByFactor(mousePosition, zoomFactor(ev));
    }
    ev.preventDefault();
  }

  clampedTranslation(state) {
    const {
      width, height, canvasMargins, boundContent, layoutLimits
    } = this.props;
    const {
      contentMinX, contentMaxX, contentMinY, contentMaxY
    } = layoutLimits.toJS();

    if (boundContent) {
      // If the content is required to be bounded in any way, the translation will
      // be adjusted so that certain constraints between the viewport and displayed
      // content bounding box are met.
      const viewportMin = { x: canvasMargins.left, y: canvasMargins.top };
      const viewportMax = { x: canvasMargins.left + width, y: canvasMargins.top + height };
      const contentMin = applyTransform(state, { x: contentMinX, y: contentMinY });
      const contentMax = applyTransform(state, { x: contentMaxX, y: contentMaxY });

      switch (boundContent) {
        case CONTENT_COVERING:
          // These lines will adjust the translation by 'minimal effort' in
          // such a way that the content always FULLY covers the viewport,
          // i.e. that the viewport rectangle is always fully contained in
          // the content bounding box rectangle - the assumption made here
          // is that that can always be done.
          state.translateX += Math.max(0, viewportMax.x - contentMax.x);
          state.translateX -= Math.max(0, contentMin.x - viewportMin.x);
          state.translateY += Math.max(0, viewportMax.y - contentMax.y);
          state.translateY -= Math.max(0, contentMin.y - viewportMin.y);
          break;
        case CONTENT_INCLUDED:
          // These lines will adjust the translation by 'minimal effort' in
          // such a way that the content is always at least PARTLY contained
          // within the viewport, i.e. that the intersection between the
          // viewport and the content bounding box always exists.
          state.translateX -= Math.max(0, contentMin.x - viewportMax.x);
          state.translateX += Math.max(0, viewportMin.x - contentMax.x);
          state.translateY -= Math.max(0, contentMin.y - viewportMax.y);
          state.translateY += Math.max(0, viewportMin.y - contentMax.y);
          break;
        default:
          break;
      }
    }

    return state;
  }

  zoomAtPositionByFactor(position, factor) {
    // Update the scales by the given factor, respecting the zoom limits.
    const { minScale, maxScale } = this.state;
    const scaleX = clamp(this.state.scaleX * factor, minScale, maxScale);
    const scaleY = clamp(this.state.scaleY * factor, minScale, maxScale);
    let state = { ...this.state, scaleX, scaleY };

    // Get the position in the coordinates before the transition and use it
    // to adjust the translation part of the new transition (respecting the
    // translation limits). Adapted from:
    // https://github.com/d3/d3-zoom/blob/807f02c7a5fe496fbd08cc3417b62905a8ce95fa/src/zoom.js#L251
    const inversePosition = inverseTransform(this.state, position);
    state = this.clampedTranslation({
      ...state,
      translateX: position.x - (inversePosition.x * scaleX),
      translateY: position.y - (inversePosition.y * scaleY),
    });

    this.updateState(state);
  }

  updateState(state) {
    this.setState(this.cachableState(state));
    this.debouncedCacheZoom();
  }
}


function mapStateToProps(state, props) {
  return {
    canvasMargins: canvasMarginsSelector(state),
    forceRelayout: state.get('forceRelayout'),
    height: canvasHeightSelector(state),
    layoutId: JSON.stringify(activeTopologyZoomCacheKeyPathSelector(state)),
    layoutLimits: props.limitsSelector(state),
    layoutZoomState: props.zoomStateSelector(state),
    width: canvasWidthSelector(state),
  };
}


export default connect(
  mapStateToProps,
  { cacheZoomState }
)(ZoomableCanvas);
