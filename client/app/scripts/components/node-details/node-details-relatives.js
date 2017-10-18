import React from 'react';
import { Map as makeMap } from 'immutable';

import { NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT } from '../../constants/limits';
import NodeDetailsRelativesLink from './node-details-relatives-link';

export default class NodeDetailsRelatives extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.state = {
      limit: NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT
    };
    this.handleLimitClick = this.handleLimitClick.bind(this);
  }

  handleLimitClick(ev) {
    ev.preventDefault();
    const limit = this.state.limit ? 0 : NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT;
    this.setState({limit});
  }

  render() {
    let { relatives } = this.props;
    const { matches = makeMap() } = this.props;

    const limited = this.state.limit > 0 && relatives.length > this.state.limit;
    const showLimitAction = limited || (this.state.limit === 0
      && relatives.length > NODE_DETAILS_DATA_ROWS_DEFAULT_LIMIT);
    const limitActionText = limited ? 'Show more' : 'Show less';
    if (limited) {
      relatives = relatives.slice(0, this.state.limit);
    }

    return (
      <div className="node-details-relatives">
        {relatives.map(relative => (
          <NodeDetailsRelativesLink
            key={relative.id}
            match={matches.get(relative.id)}
            {...relative} />))}
        {showLimitAction &&
          <span
            className="node-details-relatives-more"
            onClick={this.handleLimitClick}>
            {limitActionText}
          </span>
        }
      </div>
    );
  }
}
