import React from 'react';
import classNames from 'classnames';
import { connect } from 'react-redux';

import { doControl } from '../../actions/app-actions';


function renderButton({control, pending, optionKeyDown}, onClick) {
  const altAction = optionKeyDown && control.can_raw_pipe;
  const className = classNames('node-control-button', 'fa', control.icon, {
    'node-control-button-pending': pending,
    'node-control-button-alt': altAction
  });
  const title = altAction ? `${control.human} With RAW Pipe!` : control.human;

  return (
    <span
      className={className}
      title={title}
      onClick={onClick} />
  );
}


class NodeDetailsControlButton extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  render() {
    return (
      <span>
        {renderButton(this.props, this.handleClick)}
      </span>
    );
  }

  handleClick(ev) {
    const args = {};
    ev.preventDefault();
    if (this.props.optionKeyDown && this.props.control.can_raw_pipe) {
      args.raw_pipe = 'true';
    }
    this.props.dispatch(doControl(this.props.nodeId, this.props.control, args));
  }
}

function mapStateToProps(state) {
  return {
    optionKeyDown: state.get('optionKeyDown')
  };
}

export default connect(
  mapStateToProps
)(NodeDetailsControlButton);
