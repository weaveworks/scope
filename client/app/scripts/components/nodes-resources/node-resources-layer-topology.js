import React from 'react';
import pick from 'lodash/pick';

import { applyTransform } from '../../utils/transform-utils';
import {
  RESOURCES_LAYER_TITLE_WIDTH,
  RESOURCES_LAYER_HEIGHT,
} from '../../constants/styles';


export default class NodeResourcesLayerTopology extends React.Component {
  render() {
    // This component always has a fixed horizontal position and width,
    // so we only apply the vertical zooming transformation to match the
    // vertical position and height of the resource boxes.
    const verticalTransform = pick(this.props.transform, ['translateY', 'scaleY']);
    const { width, height, y } = applyTransform(verticalTransform, {
      height: RESOURCES_LAYER_HEIGHT,
      width: RESOURCES_LAYER_TITLE_WIDTH,
      y: this.props.verticalPosition,
    });

    return (
      <foreignObject width={width} height={height} y={y}>
        <div className="node-resources-layer-topology" style={{ lineHeight: `${height}px` }}>
          {this.props.topologyId}
        </div>
      </foreignObject>
    );
  }
}
