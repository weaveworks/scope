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
        <div style={{height: '100%', paddingBottom: 8, borderRadius: 2,
          backgroundColor: '#fff',
          boxShadow: '0 10px 30px rgba(0, 0, 0, 0.19), 0 6px 10px rgba(0, 0, 0, 0.23)'}}>
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
