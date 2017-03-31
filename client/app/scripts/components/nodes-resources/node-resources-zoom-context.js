import React from 'react';
import range from 'lodash/range';
import { connect } from 'react-redux';

import { formatMetricSvg } from '../../utils/string-utils';
import { canvasMarginsSelector, canvasWidthSelector } from '../../selectors/canvas';


const MAX_MARKERS_DISTANCE_PX = 300;
const MIN_MARKERS_DISTANCE_PX = 70;
// TODO: Ideally these shouldn't be hardcoded, but right now reading the metric format from
// the metric info of individual nodes would be both ugly and a bit incorrect (since we don't
// want a fixed precision for CPU in the zoom context, but do want it in the node info).
const METRIC_FORMATS = {
  CPU: { format: 'percent', fixedPrecision: false },
  Memory: { format: 'filesize' },
};

class NodeResourcesZoomContext extends React.Component {
  getMetricMarkers() {
    const scale = this.props.transform.scaleX;
    const minMetricDiff = MIN_MARKERS_DISTANCE_PX / scale;
    const maxMetricDiff = MAX_MARKERS_DISTANCE_PX / scale;

    // Find an appropriate power of 2 for the metric difference between
    // the two adjacent markers. We want the value to be such that the
    // markers are rendered as sparsely as possible while still being
    // at most MAX_DISTANCE_PX pixels apart from one another.
    let step = 1;
    while (step < maxMetricDiff) step *= 2;
    while (step > maxMetricDiff) step /= 2;

    // Spread the metric markers across the whole canvas width. The last
    // marker has a fixed position at the end, so we make sure that the
    // one preceding it is at least MIN_DISTANCE_PX apart from it.
    const maxMetric = this.props.canvasWidth / scale;
    const spreadLimit = maxMetric - minMetricDiff;
    return range(0, spreadLimit, step).concat(maxMetric);
  }

  renderMarker(metricValue) {
    const metricFormat = METRIC_FORMATS[this.props.pinnedMetricType];
    const formattedValue = formatMetricSvg(metricValue, metricFormat);
    const position = metricValue * this.props.transform.scaleX;
    const transform = `translate(${position}, 0)`;

    // Renders a tick on the horizontal zoom scale,
    // together with the formatted metric value.
    return (
      <g className="zoom-context-marker" key={metricValue} transform={transform}>
        <line className="zoom-context-line" y1="-10" />
        <foreignObject x="-30" y="5" width="60" height="20">
          <span className="zoom-context-marker">{formattedValue}</span>
        </foreignObject>
      </g>
    );
  }

  render() {
    const translateY = this.props.transform.translateY + 20;
    const { canvasWidth, canvasMarginLeft, pinnedMetricType } = this.props;
    const transform = `translate(${canvasMarginLeft}, ${translateY})`;

    return (
      <g className="zoom-context" transform={transform}>
        <foreignObject className="zoom-context-label" x="-115" y="-10" width="100" height="20">
          <span>{pinnedMetricType}</span>
        </foreignObject>
        <line className="zoom-context-line" x2={canvasWidth} />
        {this.getMetricMarkers().map(value => this.renderMarker(value))}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    canvasWidth: canvasWidthSelector(state),
    canvasMarginLeft: canvasMarginsSelector(state).left,
    pinnedMetricType: state.get('pinnedMetricType'),
  };
}

export default connect(mapStateToProps)(NodeResourcesZoomContext);
