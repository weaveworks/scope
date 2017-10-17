import React from 'react';
import { connect } from 'react-redux';

import { unpinSearch } from '../actions/app-actions';

class SearchItem extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.unpinSearch(this.props.query);
  }

  render() {
    return (
      <span className="search-item">
        <span className="search-item-label">{this.props.query}</span>
        <span className="search-item-icon fa fa-close" onClick={this.handleClick} />
      </span>
    );
  }
}

export default connect(null, { unpinSearch })(SearchItem);
