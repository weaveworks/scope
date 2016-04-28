import React from 'react';
import { connect } from 'react-redux';
import cx from 'classnames';
import _ from 'lodash';

import { blurSearch, doSearch, focusSearch } from '../actions/app-actions';
import { isTopologyEmpty } from '../utils/topology-utils';

const SEARCH_HINTS = [
  'Try "db" or "app1" to search by node label or sublabel.',
  'Try "sublabel:my-host" to search by node sublabel.',
  'Try "label:my-node" to search by node name.',
  'Try "metadata:my-node" to search through all node metadata.',
  'Try "dockerenv:my-env-value" to search through all docker environment variables.',
  'Try "all:my-value" to search through all metdata and labels.'
];

// every minute different hint
function getHint() {
  return SEARCH_HINTS[(new Date).getMinutes() % SEARCH_HINTS.length];
}

class Search extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleBlur = this.handleBlur.bind(this);
    this.handleChange = this.handleChange.bind(this);
    this.handleFocus = this.handleFocus.bind(this);
    this.doSearch = _.debounce(this.doSearch.bind(this), 200);
    this.state = {
      value: ''
    };
  }

  handleBlur() {
    this.props.blurSearch();
  }

  handleChange(ev) {
    const value = ev.target.value;
    this.setState({value});
    this.doSearch(value);
  }

  handleFocus() {
    this.props.focusSearch();
  }

  doSearch(value) {
    this.props.doSearch(value);
  }

  componentWillReceiveProps(nextProps) {
    // when cleared from the outside, reset internal state
    if (this.props.searchQuery !== nextProps.searchQuery && nextProps.searchQuery === '') {
      this.setState({value: ''});
    }
  }

  render() {
    const inputId = this.props.inputId || 'search';
    const disabled = this.props.isTopologyEmpty || !this.props.topologiesLoaded;
    const matchCount = this.props.searchNodeMatches
      .reduce((count, topologyMatches) => count + topologyMatches.size, 0);
    const classNames = cx('search', {
      'search-matched': matchCount,
      'search-filled': this.state.value,
      'search-focused': this.props.searchFocused,
      'search-disabled': disabled
    });
    const title = matchCount ? `${matchCount} matches` : null;

    return (
      <div className="search-wrapper">
        <div className={classNames} title={title}>
          <div className="search-input">
            <input className="search-input-field" type="text" id={inputId}
              value={this.state.value} onChange={this.handleChange}
              onBlur={this.handleBlur} onFocus={this.handleFocus}
              disabled={disabled} />
            <label className="search-input-label" htmlFor={inputId}>
              <i className="fa fa-search search-input-label-icon"></i>
              <span className="search-input-label-text">Search</span>
            </label>
          </div>
          <div className="search-hint">{getHint()}</div>
        </div>
      </div>
    );
  }
}

export default connect(
  state => ({
    isTopologyEmpty: isTopologyEmpty(state),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    searchNodeMatches: state.get('searchNodeMatches'),
    topologiesLoaded: state.get('topologiesLoaded')
  }),
  { blurSearch, doSearch, focusSearch }
)(Search);
