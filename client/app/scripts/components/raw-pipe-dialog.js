import React from 'react';
import { connect } from 'react-redux';
import { clickCloseTerminal } from '../actions/app-actions';


class RawPipeDialog extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClickClose = this.handleClickClose.bind(this);
  }

  handleClickClose() {
    this.props.dispatch(clickCloseTerminal(this.props.controlPipes.first().get('id'), false));
  }

  render() {
    const {controlPipes} = this.props;
    const pipeDetails = controlPipes.first().get('rawPipeTemplate');
    return (
      <div className="help-panel">
        <div className="help-panel-header">
          <h2>Raw Pipe Opened!</h2>
        <span title="Close details" className="fa fa-close" onClick={this.handleClickClose} />
        </div>
        <div className="help-panel-main">
          {pipeDetails}
        </div>
      </div>
    );
  }
}


function mapStateToProps(state) {
  return {
    controlPipes: state.get('controlPipes')
  };
}


export default connect(
  mapStateToProps
)(RawPipeDialog);
