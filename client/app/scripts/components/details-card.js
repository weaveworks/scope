import React from 'react';

import NodeDetails from './node-details';

// card dimensions in px
const marginTop = 24;
const marginBottom = 48;
const marginRight = 36;
const panelWidth = 420;
const offset = 8;

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
    const panelHeight = window.innerHeight - marginBottom - marginTop;
    if (origin && !this.state.mounted) {
      // render small panel near origin, will transition into normal panel after being mounted
      const scaleY = origin.height / (window.innerHeight - marginBottom - marginTop) / 2;
      const scaleX = origin.width / panelWidth / 2;
      const centerX = window.innerWidth - marginRight - (panelWidth / 2);
      const centerY = (panelHeight) / 2 + marginTop;
      const dx = (origin.left + origin.width / 2) - centerX;
      const dy = (origin.top + origin.height / 2) - centerY;
      transform = `translate(${dx}px, ${dy}px) scale(${scaleX},${scaleY})`;
    } else {
      // stack effect: shift top cards to the left, shrink lower cards vertically
      const shiftX = -1 * this.props.index * offset;
      const position = this.props.cardCount - this.props.index - 1; // reverse index
      const scaleY = position === 0 ? 1 : (panelHeight - 2 * offset * position) / panelHeight;
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
