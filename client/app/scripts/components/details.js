import React from 'react';

import { clickCloseDetails } from '../actions/app-actions';
import NodeDetails from './node-details';

export default class Details extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClickClose = this.handleClickClose.bind(this);
  }

  handleClickClose(ev) {
    ev.preventDefault();
    clickCloseDetails();
  }

  render() {
    return (
      <div id="details">
        <div className="details-wrapper">
          <div className="details-tools-wrapper">
            <div className="details-tools">
              <span className="fa fa-close" onClick={this.handleClickClose} />
            </div>
          </div>
          <NodeDetails {...this.props} />
        </div>
      </div>
    );
  }
}
