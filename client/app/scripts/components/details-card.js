import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import NodeDetails from './node-details';
import { DETAILS_PANEL_WIDTH as WIDTH, DETAILS_PANEL_OFFSET as OFFSET,
  DETAILS_PANEL_MARGINS as MARGINS } from '../constants/styles';

export default class DetailsCard extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.state = {
      mounted: null
    };
  }

  componentDidMount() {
    setTimeout(() => {
      this.setState({mounted: true});
    });
  }

  render() {
    let transform;
    const origin = this.props.origin;
    const panelHeight = window.innerHeight - MARGINS.bottom - MARGINS.top;
    if (origin && !this.state.mounted) {
      // render small panel near origin, will transition into normal panel after being mounted
      const scaleY = origin.height / (window.innerHeight - MARGINS.bottom - MARGINS.top) / 2;
      const scaleX = origin.width / WIDTH / 2;
      const centerX = window.innerWidth - MARGINS.right - (WIDTH / 2);
      const centerY = (panelHeight) / 2 + MARGINS.top;
      const dx = (origin.left + origin.width / 2) - centerX;
      const dy = (origin.top + origin.height / 2) - centerY;
      transform = `translate(${dx}px, ${dy}px) scale(${scaleX},${scaleY})`;
    } else {
      // stack effect: shift top cards to the left, shrink lower cards vertically
      const shiftX = -1 * this.props.index * OFFSET;
      const position = this.props.cardCount - this.props.index - 1; // reverse index
      const scaleY = position === 0 ? 1 : (panelHeight - 2 * OFFSET * position) / panelHeight;
      if (scaleY !== 1) {
        transform = `translateX(${shiftX}px) scaleY(${scaleY})`;
      } else {
        // scale(1) is sometimes blurry
        transform = `translateX(${shiftX}px)`;
      }
    }
    return (
      <div className="details-wrapper" style={{transform}}>
        <NodeDetails nodeId={this.props.id} key={this.props.id} {...this.props} />
      </div>
    );
  }
}

reactMixin.onClass(DetailsCard, PureRenderMixin);
