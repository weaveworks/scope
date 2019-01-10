import React from 'react';
import { connect } from 'react-redux';

import { brightenColor, getNodeColorDark } from '../utils/color-utils';
import { DETAILS_PANEL_WIDTH, DETAILS_PANEL_MARGINS } from '../constants/styles';
import Terminal from './terminal';

class EmeddedTerminal extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      animated: null,
      mounted: null,
    };

    this.handleTransitionEnd = this.handleTransitionEnd.bind(this);
  }

  componentDidMount() {
    this.mountedTimeout = setTimeout(() => {
      this.setState({mounted: true});
    });
    this.animationTimeout = setTimeout(() => {
      this.setState({ animated: true });
    }, 2000);
  }

  componentWillUnmount() {
    clearTimeout(this.mountedTimeout);
    clearTimeout(this.animationTimeout);
  }

  getTransform() {
    const dx = this.state.mounted ? 0 :
      window.innerWidth - DETAILS_PANEL_WIDTH - DETAILS_PANEL_MARGINS.right;
    return `translateX(${dx}px)`;
  }

  handleTransitionEnd() {
    this.setState({ animated: true });
  }

  render() {
    const { pipe, details } = this.props;
    const nodeId = pipe.get('nodeId');
    const node = details.get(nodeId);
    const d = node && node.details;
    const titleBarColor = d && getNodeColorDark(d.rank, d.label, d.pseudo);
    const statusBarColor = d && brightenColor(titleBarColor);
    const title = d && d.label;

    // React unmount/remounts when key changes, this is important for cleaning up
    // the term.js and creating a new one for the new pipe.
    return (
      <div className="tour-step-anchor terminal-embedded">
        <div
          onTransitionEnd={this.handleTransitionEnd}
          className="terminal-animation-wrapper"
          style={{transform: this.getTransform()}} >
          <Terminal
            key={pipe.get('id')}
            pipe={pipe}
            connect={this.state.animated}
            titleBarColor={titleBarColor}
            statusBarColor={statusBarColor}
            title={title} />
        </div>
      </div>
    );
  }
}

function mapStateToProps(state) {
  return {
    details: state.get('nodeDetails'),
    pipe: state.get('controlPipes').last()
  };
}

export default connect(mapStateToProps)(EmeddedTerminal);
