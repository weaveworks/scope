import React from 'react';
import { fromJS } from 'immutable';

import NodeResourcesMetricBoxInfo from './node-resources-metric-box-info';
import { applyTransformX, applyTransformY } from '../../utils/transform-utils';
import {
  RESOURCES_LAYER_TITLE_WIDTH,
  RESOURCES_LABEL_MIN_SIZE,
  RESOURCES_LABEL_PADDING,
} from '../../constants/styles';


export default class LayerLabelsOverlay extends React.Component {
  positionedLabels() {
    const { verticalOffset, transform, nodes } = this.props;
    const y = applyTransformY(transform, verticalOffset);
    const labels = [];

    nodes.forEach((node) => {
      const xStart = applyTransformX(transform, node.get('offset'));
      const xEnd = applyTransformX(transform, node.get('offset') + node.get('width'));
      const xTrimmed = Math.max(RESOURCES_LAYER_TITLE_WIDTH, xStart);
      const width = xEnd - xTrimmed;

      if (width >= RESOURCES_LABEL_MIN_SIZE) {
        labels.push({
          width: width - (2 * RESOURCES_LABEL_PADDING),
          x: xTrimmed + RESOURCES_LABEL_PADDING,
          y: y + RESOURCES_LABEL_PADDING,
          node,
        });
      }
    });

    return fromJS(labels);
  }

  render() {
    return (
      <g className="labels-overlay">
        {this.positionedLabels().map(label => (
          <NodeResourcesMetricBoxInfo
            key={label.getIn(['node', 'id'])}
            node={label.get('node')}
            width={label.get('width')}
            x={label.get('x')}
            y={label.get('y')}
          />
        ))}
      </g>
    );
  }
}
