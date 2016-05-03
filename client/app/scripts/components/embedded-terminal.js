import React from 'react';
import { connect } from 'react-redux';

import { getNodeColor, getNodeColorDark } from '../utils/color-utils';
import Terminal from './terminal';
import { DETAILS_PANEL_WIDTH, DETAILS_PANEL_MARGINS,
  DETAILS_PANEL_OFFSET } from '../constants/styles';

class EmeddedTerminal extends React.Component {
  render() {
    const { pipe, details } = this.props;
    const nodeId = pipe.get('nodeId');
    const node = details.get(nodeId);
    const d = node && node.details;
    const titleBarColor = d && getNodeColorDark(d.rank, d.label);
    const statusBarColor = d && getNodeColor(d.rank, d.label);
    const title = d && d.label;

    const style = {
      right: DETAILS_PANEL_MARGINS.right + DETAILS_PANEL_WIDTH + 10 +
        (details.size * DETAILS_PANEL_OFFSET)
    };

    // React unmount/remounts when key changes, this is important for cleaning up
    // the term.js and creating a new one for the new pipe.
    return (
      <div className="terminal-embedded" style={style}>
        <Terminal key={pipe.get('id')} pipe={pipe} titleBarColor={titleBarColor}
          statusBarColor={statusBarColor} containerMargin={style.right}
          title={title} />
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

export default connect(
  mapStateToProps
)(EmeddedTerminal);
