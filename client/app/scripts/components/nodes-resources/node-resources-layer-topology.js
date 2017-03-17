import React from 'react';

import { RESOURCES_LAYER_TITLE_WIDTH, RESOURCES_LAYER_HEIGHT } from '../../constants/styles';
import { applyTransformY } from '../../utils/transform-utils';

export default class LayerTopologyName extends React.Component {
  render() {
    const { verticalOffset, topologyId, transform } = this.props;
    const height = RESOURCES_LAYER_HEIGHT * transform.scaleY;
    const y = applyTransformY(transform, verticalOffset);

    return (
      <foreignObject
        className="layer-topology-name"
        width={RESOURCES_LAYER_TITLE_WIDTH}
        height={height}
        y={y}>
        <span style={{ height, lineHeight: `${height}px` }}>{topologyId}</span>
      </foreignObject>
    );
  }
}
