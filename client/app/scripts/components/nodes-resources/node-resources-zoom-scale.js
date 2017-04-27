import React from 'react';
import { connect } from 'react-redux';

import { formatMetricSvg } from '../../utils/string-utils';
import { transformToString } from '../../utils/transform-utils';
import { greatestPowerOfTwoNotExceeding } from '../../utils/math-utils';
import { canvasMarginsSelector } from '../../selectors/canvas';


// TODO: Ideally these shouldn't be hardcoded, but right now reading the metric format from
// the metric info of individual nodes would be both ugly and a bit incorrect (since we don't
// want a fixed precision for CPU in the zoom context, but do want it in the node info).
const METRIC_FORMATS = {
  CPU: { format: 'percent', fixedPrecision: false },
  Memory: { format: 'filesize' },
};
const MAX_SCALE_LENGTH = 200; // in pixels


class NodeResourcesZoomScale extends React.Component {
  getScaleParameters() {
    const metricFormat = METRIC_FORMATS[this.props.pinnedMetricType];
    // Find the metric value that would correspond to showing the scale at the maximal length.
    const maxMetricValue = MAX_SCALE_LENGTH / this.props.zoomLevel;
    // Find the biggest power of 2 metric value not exceeding it
    // - that will be the one we will be showing on the scale.
    const metricValue = greatestPowerOfTwoNotExceeding(maxMetricValue);

    return {
      // Multiply by the zoom level to get the actual scale length back.
      width: metricValue * this.props.zoomLevel,
      formattedMetricValue: formatMetricSvg(metricValue, metricFormat),
    };
  }

  render() {
    const { viewportWidth, viewportHeight, canvasMarginRight } = this.props;
    const { width, formattedMetricValue } = this.getScaleParameters();

    // Position the zoom scale just above the footer.
    const transform = transformToString({
      translateX: viewportWidth - canvasMarginRight - width - 12,
      translateY: viewportHeight - 50,
    });

    return (
      <g className="nodes-resources-zoom-scale" transform={transform}>
        <path className="frame" d={`M 0 -15 V 0 H ${width} V -15`} />
        <foreignObject className="value" y="-20" height="20" width={width}>
          <span>{formattedMetricValue}</span>
        </foreignObject>
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    canvasMarginRight: canvasMarginsSelector(state).right,
    viewportWidth: state.getIn(['viewport', 'width']),
    viewportHeight: state.getIn(['viewport', 'height']),
    pinnedMetricType: state.get('pinnedMetricType'),
  };
}

export default connect(mapStateToProps)(NodeResourcesZoomScale);
