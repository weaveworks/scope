import React from 'react';
import { connect } from 'react-redux';
import theme from 'weaveworks-ui-components/lib/theme';

import NodeResourcesMetricBoxInfo from './node-resources-metric-box-info';
import { clickNode } from '../../actions/app-actions';
import { trackAnalyticsEvent } from '../../utils/tracking-utils';
import { applyTransform } from '../../utils/transform-utils';
import { RESOURCE_VIEW_MODE } from '../../constants/naming';
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
  const {
    width, height, x, y
  } = applyTransform(props.transform, props);

  // Trim the beginning of the resource box just after the layer topology
  // name to the left and the viewport width to the right. That enables us
  // to make info tags 'sticky', but also not to render the nodes with no
  // visible part in the viewport.
  const xStart = Math.max(RESOURCES_LAYER_TITLE_WIDTH, x);
  const xEnd = Math.min(x + width, props.viewportWidth);

  // Update the horizontal transform with trimmed values.
  return {
    height,
    width: xEnd - xStart,
    x: xStart,
    y,
  };
};

class NodeResourcesMetricBox extends React.Component {
  constructor(props, context) {
    super(props, context);

    this.state = transformedDimensions(props);

    this.handleClick = this.handleClick.bind(this);
    this.saveNodeRef = this.saveNodeRef.bind(this);
  }

  componentWillReceiveProps(nextProps) {
    this.setState(transformedDimensions(nextProps));
  }

  handleClick(ev) {
    ev.stopPropagation();
    trackAnalyticsEvent('scope.node.click', {
      layout: RESOURCE_VIEW_MODE,
      topologyId: this.props.topologyId,
    });
    this.props.clickNode(
      this.props.id,
      this.props.label,
      this.nodeRef.getBoundingClientRect(),
      this.props.topologyId
    );
  }

  saveNodeRef(ref) {
    this.nodeRef = ref;
  }

  defaultRectProps(relativeHeight = 1) {
    const {
      x, y, width, height
    } = this.state;
    const translateY = height * (1 - relativeHeight);
    return {
      height: height * relativeHeight,
      opacity: this.props.contrastMode ? 1 : 0.85,
      stroke: this.props.contrastMode ? 'black' : 'white',
      transform: `translate(0, ${translateY})`,
      width,
      x,
      y,
    };
  }

  render() {
    const { x, y, width } = this.state;
    const {
      id, selectedNodeId, label, color, metricSummary
    } = this.props;
    const { showCapacity, relativeConsumption, type } = metricSummary.toJS();
    const opacity = (selectedNodeId && selectedNodeId !== id) ? 0.35 : 1;

    const showInfo = width >= RESOURCES_LABEL_MIN_SIZE;
    const showNode = width >= 1; // hide the thin nodes

    // Don't display the nodes which are less than 1px wide.
    // TODO: Show `+ 31 nodes` kind of tag in their stead.
    if (!showNode) return null;

    const resourceUsageTooltipInfo = showCapacity ?
      metricSummary.get('humanizedRelativeConsumption') :
      metricSummary.get('humanizedAbsoluteConsumption');

    return (
      <g
        className="node-resources-metric-box"
        style={{ opacity }}
        onClick={this.handleClick}
        ref={this.saveNodeRef}
      >
        <title>{label} - {type} usage at {resourceUsageTooltipInfo}</title>
        {showCapacity && <rect
          className="frame"
          rx={theme.borderRadius.soft}
          ry={theme.borderRadius.soft}
          {...this.defaultRectProps()}
        />}
        <rect
          className="bar"
          fill={color}
          rx={theme.borderRadius.soft}
          ry={theme.borderRadius.soft}
          {...this.defaultRectProps(relativeConsumption)}
        />
        {showInfo && <NodeResourcesMetricBoxInfo
          label={label}
          metricSummary={metricSummary}
          width={width - (2 * RESOURCES_LABEL_PADDING)}
          x={x + RESOURCES_LABEL_PADDING}
          y={y + RESOURCES_LABEL_PADDING}
        />}
      </g>
    );
  }
}

function mapStateToProps(state) {
  return {
    contrastMode: state.get('contrastMode'),
    selectedNodeId: state.get('selectedNodeId'),
    viewportWidth: state.getIn(['viewport', 'width']),
  };
}
export default connect(
  mapStateToProps,
  { clickNode }
)(NodeResourcesMetricBox);
