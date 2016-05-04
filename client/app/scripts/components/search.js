import React from 'react';
import { connect } from 'react-redux';
import cx from 'classnames';
import _ from 'lodash';

import { blurSearch, doSearch, focusSearch } from '../actions/app-actions';
import { slugify } from '../utils/string-utils';
import { isTopologyEmpty } from '../utils/topology-utils';
import SearchItem from './search-item';

// dynamic hint based on node names
function getHint(nodes) {
  let label = 'mycontainer';
  let metadataLabel = 'ip';
  let metadataValue = '172.12';

  const node = nodes.last();
  if (node) {
    label = node.get('label');
    if (node.get('metadata')) {
      const metadataField = node.get('metadata').first();
      metadataLabel = slugify(metadataField.get('label'))
        .split(' ')[0]
        .split('.').pop()
        .substr(0, 20);
      metadataValue = metadataField.get('value')
        .toLowerCase()
        .split(' ')[0]
        .substr(0, 12);
    }
  }

  return `Try "${label}" or "${metadataLabel}:${metadataValue}".
   Hit enter to apply the search as a filter.`;
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
    const { inputId = 'search', nodes, pinnedSearches, searchFocused,
      searchNodeMatches, topologiesLoaded } = this.props;
    const disabled = this.props.isTopologyEmpty || !topologiesLoaded;
    const matchCount = searchNodeMatches
      .reduce((count, topologyMatches) => count + topologyMatches.size, 0);
    const showPinnedSearches = pinnedSearches.size > 0;
    const classNames = cx('search', {
      'search-pinned': showPinnedSearches,
      'search-matched': matchCount,
      'search-filled': this.state.value,
      'search-focused': searchFocused,
      'search-disabled': disabled
    });
    const title = matchCount ? `${matchCount} matches` : null;

    return (
      <div className="search-wrapper">
        <div className={classNames} title={title}>
          <div className="search-input">
            <i className="fa fa-search search-input-icon"></i>
            <label className="search-input-label" htmlFor={inputId}>
              Search
            </label>
            {showPinnedSearches && <span className="search-input-items">
              {pinnedSearches.toIndexedSeq()
                .map(query => <SearchItem query={query} key={query} />)}
            </span>}
            <input className="search-input-field" type="text" id={inputId}
              value={this.state.value} onChange={this.handleChange}
              onBlur={this.handleBlur} onFocus={this.handleFocus}
              disabled={disabled} />
          </div>
          {!showPinnedSearches && <div className="search-hint">
            {getHint(nodes)}
          </div>}
        </div>
      </div>
    );
  }
}

export default connect(
  state => ({
    nodes: state.get('nodes'),
    isTopologyEmpty: isTopologyEmpty(state),
    pinnedSearches: state.get('pinnedSearches'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    searchNodeMatches: state.get('searchNodeMatches'),
    topologiesLoaded: state.get('topologiesLoaded')
  }),
  { blurSearch, doSearch, focusSearch }
)(Search);
