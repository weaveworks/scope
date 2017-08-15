import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';
import { clamp, debounce, pick } from 'lodash';
import { fromJS } from 'immutable';

import { drag } from 'd3-drag';
import { event as d3Event, select } from 'd3-selection';

import Logo from '../components/logo';
import ZoomControl from '../components/zoom-control';
import { cacheZoomState } from '../actions/app-actions';
import { zoomFactor } from '../utils/zoom-utils';
import { applyTransform, inverseTransform } from '../utils/transform-utils';
import { activeTopologyZoomCacheKeyPathSelector } from '../selectors/zooming';
import {
  canvasMarginsSelector,
  canvasWidthSelector,
  canvasHeightSelector,
} from '../selectors/canvas';

import { ZOOM_CACHE_DEBOUNCE_INTERVAL } from '../constants/timer';


class ZoomableCanvas extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = {
      isPanning: false,
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
  }

  componentWillUnmount() {
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

    this.updateZoomLimits(nextProps);
    if (!this.zoomRestored) {
      this.restoreZoomState(nextProps);
    }
  }

  handleZoomControlAction(scale) {
    // Update the canvas scale (not touching the translation).
    const { top, bottom, left, right } = this.svg.node().getBoundingClientRect();
    const centerOfCanvas = {
      x: (left + right) / 2,
      y: (top + bottom) / 2,
    };
    this.zoomAtPosition(centerOfCanvas, scale / this.state.scaleX);
  }

  render() {
    const className = classNames({ panning: this.state.isPanning });

    return (
      <div className="zoomable-canvas">
        <svg className={className} onClick={this.props.onClick} onWheel={this.handleZoom}>
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
    this.setState(props.layoutZoomLimits.toJS());
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
    const { disabled, layoutZoomLimits } = this.props;
    const canvasHasContent = !layoutZoomLimits.isEmpty();
    return !disabled && canvasHasContent;
  }

  handlePanStart() {
    this.setState({ isPanning: true });
  }

  handlePanEnd() {
    this.setState({ isPanning: false });
  }

  handlePan() {
    let state = { ...this.state };
    state = this.clampedTranslation({ ...state,
      translateX: this.state.translateX + d3Event.dx,
      translateY: this.state.translateY + d3Event.dy,
    });
    this.updateState(state);
  }

  handleZoom(ev) {
    if (this.canChangeZoom()) {
      const { top, left } = this.svg.node().getBoundingClientRect();
      const mousePosition = {
        x: ev.clientX - left,
        y: ev.clientY - top,
      };
      this.zoomAtPosition(mousePosition, 1 / zoomFactor(ev));
    }
  }

  clampedTranslation(state) {
    const { width, height, canvasMargins, bounded, layoutZoomLimits } = this.props;
    const { maxTranslateX, minTranslateX, maxTranslateY, minTranslateY } = layoutZoomLimits.toJS();

    if (bounded) {
      const contentMinPoint = applyTransform(state, { x: minTranslateX, y: minTranslateY });
      const contentMaxPoint = applyTransform(state, { x: maxTranslateX, y: maxTranslateY });
      const viewportMinPoint = { x: canvasMargins.left, y: canvasMargins.top };
      const viewportMaxPoint = { x: canvasMargins.left + width, y: canvasMargins.top + height };

      state.translateX += Math.max(0, viewportMaxPoint.x - contentMaxPoint.x);
      state.translateX += Math.min(0, viewportMinPoint.x - contentMinPoint.x);
      state.translateY += Math.max(0, viewportMaxPoint.y - contentMaxPoint.y);
      state.translateY += Math.min(0, viewportMinPoint.y - contentMinPoint.y);
    }

    return state;
  }

  zoomAtPosition(position, factor) {
    const { minScale, maxScale } = this.state;
    const scaleX = clamp(this.state.scaleX * factor, minScale, maxScale);
    const scaleY = clamp(this.state.scaleY * factor, minScale, maxScale);
    let state = { ...this.state, scaleX, scaleY };

    const inversePosition = inverseTransform(this.state, position);
    state = this.clampedTranslation({ ...state,
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
    width: canvasWidthSelector(state),
    height: canvasHeightSelector(state),
    canvasMargins: canvasMarginsSelector(state),
    layoutZoomState: props.zoomStateSelector(state),
    layoutZoomLimits: props.zoomLimitsSelector(state),
    layoutId: JSON.stringify(activeTopologyZoomCacheKeyPathSelector(state)),
    forceRelayout: state.get('forceRelayout'),
  };
}


export default connect(
  mapStateToProps,
  { cacheZoomState }
)(ZoomableCanvas);
