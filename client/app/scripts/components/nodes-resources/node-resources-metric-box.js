import React from 'react';
import { connect } from 'react-redux';

import NodeResourcesMetricBoxInfo from './node-resources-metric-box-info';
import { applyTransform } from '../../utils/transform-utils';
import {
  RESOURCES_LAYER_TITLE_WIDTH,
  RESOURCES_LABEL_MIN_SIZE,
  RESOURCES_LABEL_PADDING,
} from '../../constants/styles';


// Transforms the rectangle box according to the zoom state forwarded by
// the zooming wrapper. Two main reasons why we're doing it per component
// instead of on the parent group are:
//   1. Due to single-precision SVG coordinate system implemented by most browsers,
//      the resource boxes would be incorrectly rendered on extreme zoom levels (it's
//      not just about a few pixels here and there, the whole layout gets screwed). So
//      we don't actually use the native SVG transform but transform the coordinates
//      ourselves (with `applyTransform` helper).
//   2. That also enables us to do the resources info label clipping, which would otherwise
//      not be possible with pure zooming.
//
// The downside is that the rendering becomes slower as the transform prop needs to be forwarded
// down to this component, so a lot of stuff gets rerendered/recalculated on every zoom action.
// On the other hand, this enables us to easily leave out the nodes that are not in the viewport.
const transformedDimensions = (props) => {
  const { width, height, x, y } = applyTransform(props.transform, props);

  // Trim the beginning of the resource box just after the layer topology
  // name to the left and the viewport width to the right. That enables us
  // to make info tags 'sticky', but also not to render the nodes with no
  // visible part in the viewport.
  const xStart = Math.max(RESOURCES_LAYER_TITLE_WIDTH, x);
  const xEnd = Math.min(x + width, props.viewportWidth);

  // Update the horizontal transform with trimmed values.
  return {
    width: xEnd - xStart,
    height,
    x: xStart,
    y,
  };
};

class NodeResourcesMetricBox extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = transformedDimensions(props);
  }

  componentWillReceiveProps(nextProps) {
    this.setState(transformedDimensions(nextProps));
  }

  defaultRectProps(relativeHeight = 1) {
    const { x, y, width, height } = this.state;
    const translateY = height * (1 - relativeHeight);
    return {
      transform: `translate(0, ${translateY})`,
      opacity: this.props.contrastMode ? 1 : 0.85,
      // stroke: this.props.contrastMode ? 'black' : 'white',
      stroke: this.props.color,
      strokeWidth: 0.5,
      height: height * relativeHeight,
      width,
      x,
      y,
    };
  }

  render() {
    const { x, y, width } = this.state;
    const { label, fill, metricSummary, topId } = this.props;
    const { showCapacity, type } = metricSummary.toJS();

    let t;
    if (topId === 'hosts') t = 'host';
    if (topId === 'containers') t = 'container';
    if (topId === 'processes') t = 'process';

    const showInfo = width >= RESOURCES_LABEL_MIN_SIZE;
    const showNode = width >= 1; // hide the thin nodes

    // Don't display the nodes which are less than 1px wide.
    // TODO: Show `+ 31 nodes` kind of tag in their stead.
    if (!showNode) return null;

    const resourceUsageTooltipInfo = showCapacity ?
      metricSummary.get('humanizedRelativeConsumption') :
      metricSummary.get('humanizedAbsoluteConsumption');

    return (
      <g className="node-resources-metric-box">
        <title>{label} - {type} usage at {resourceUsageTooltipInfo}</title>
        <rect className="frame" {...this.defaultRectProps()} />
        <rect className="bar" fill={fill} {...this.defaultRectProps(1)} />
        {showInfo && <NodeResourcesMetricBoxInfo
          label={label}
          metricSummary={metricSummary}
          width={width - (2 * RESOURCES_LABEL_PADDING)}
          x={x + RESOURCES_LABEL_PADDING}
          y={y + RESOURCES_LABEL_PADDING}
          type={t}
        />}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode'),
    viewportWidth: state.getIn(['viewport', 'width']),
  };
}

export default connect(mapStateToProps)(NodeResourcesMetricBox);
