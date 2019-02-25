import React from 'react';
import { connect } from 'react-redux';
import { isEmpty } from 'lodash';
import classNames from 'classnames';

import { trackAnalyticsEvent } from '../../utils/tracking-utils';
import { doControl } from '../../actions/app-actions';

class NodeDetailsControlButton extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  render() {
    const { icon, id, human } = this.props.control;
    const className = classNames('tour-step-anchor node-control-button', icon, {
      // Old Agent / plugins don't include the 'fa ' prefix, so provide it if they don't.
      fa: icon.startsWith('fa-'),
      'node-control-button-pending': this.props.pending
    });
    return (
      <i className={className} data-id={id} title={human} onClick={this.handleClick} />
    );
  }

  handleClick(ev) {
    ev.preventDefault();
    const { id, human, confirmation } = this.props.control;
    trackAnalyticsEvent('scope.node.control.click', { id, title: human });
    if (isEmpty(confirmation) || window.confirm(confirmation)) { // eslint-disable-line no-alert
      this.props.dispatch(doControl(this.props.nodeId, this.props.control));
    }
  }
}

// Using this instead of PureComponent because of props.dispatch
export default connect()(NodeDetailsControlButton);
