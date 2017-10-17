import React from 'react';
import { connect } from 'react-redux';
import classnames from 'classnames';
import { debounce } from 'lodash';

import { blurSearch, doSearch, focusSearch, pinSearch, toggleHelp } from '../actions/app-actions';
import { searchMatchCountByTopologySelector } from '../selectors/search';
import { isResourceViewModeSelector } from '../selectors/topology';
import { slugify } from '../utils/string-utils';
import { parseQuery } from '../utils/search-utils';
import { isTopologyNodeCountZero } from '../utils/topology-utils';
import { trackAnalyticsEvent } from '../utils/tracking-utils';
import SearchItem from './search-item';
import { ENTER_KEY_CODE } from '../constants/key-codes';


function shortenHintLabel(text) {
  return text
    .split(' ')[0]
    .toLowerCase()
    .substr(0, 12);
}


// dynamic hint based on node names
function getHint(nodes) {
  let label = 'mycontainer';
  let metadataLabel = 'ip';
  let metadataValue = '10.1.0.1';

  const node = nodes.filter(n => !n.get('pseudo') && n.has('metadata')).last();
  if (node) {
    label = shortenHintLabel(node.get('label'))
      .split('.')[0];
    if (node.get('metadata')) {
      const metadataField = node.get('metadata').first();
      metadataLabel = shortenHintLabel(slugify(metadataField.get('label')))
        .split('.').pop();
      metadataValue = shortenHintLabel(metadataField.get('value'));
    }
  }

  return `Try "${label}", "${metadataLabel}:${metadataValue}", or "cpu > 2%".
   Hit enter to apply the search as a filter.`;
}


class Search extends React.Component {
  constructor(props, context) {
    super(props, context);
    this.handleBlur = this.handleBlur.bind(this);
    this.handleChange = this.handleChange.bind(this);
    this.handleKeyUp = this.handleKeyUp.bind(this);
    this.handleFocus = this.handleFocus.bind(this);
    this.saveQueryInputRef = this.saveQueryInputRef.bind(this);
    this.doSearch = debounce(this.doSearch.bind(this), 200);
    this.state = {
      value: ''
    };
  }

  handleBlur() {
    this.props.blurSearch();
  }

  handleChange(ev) {
    const inputValue = ev.target.value;
    let value = inputValue;
    // In render() props.searchQuery can be set from the outside, but state.value
    // must have precedence for quick feedback. Now when the user backspaces
    // quickly enough from `text`, a previouse doSearch(`text`) will come back
    // via props and override the empty state.value. To detect this edge case
    // we instead set value to null when backspacing.
    if (this.state.value && value === '') {
      value = null;
    }
    this.setState({ value });
    this.doSearch(inputValue);
  }

  handleKeyUp(ev) {
    // If the search query is parsable, pin it when ENTER key is hit.
    if (ev.keyCode === ENTER_KEY_CODE && parseQuery(this.props.searchQuery)) {
      trackAnalyticsEvent('scope.search.query.pin', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
      this.props.pinSearch();
    }
  }

  handleFocus() {
    this.props.focusSearch();
  }

  doSearch(value) {
    if (value !== '') {
      trackAnalyticsEvent('scope.search.query.change', {
        layout: this.props.topologyViewMode,
        topologyId: this.props.currentTopology.get('id'),
        parentTopologyId: this.props.currentTopology.get('parentId'),
      });
    }
    this.props.doSearch(value);
  }

  saveQueryInputRef(ref) {
    this.queryInput = ref;
  }

  componentWillReceiveProps(nextProps) {
    // when cleared from the outside, reset internal state
    if (this.props.searchQuery !== nextProps.searchQuery && nextProps.searchQuery === '') {
      this.setState({ value: '' });
    }
  }

  componentDidUpdate() {
    if (this.props.searchFocused) {
      this.queryInput.focus();
    } else if (!this.state.value) {
      this.queryInput.blur();
    }
  }

  render() {
    const {
      nodes, pinnedSearches, searchFocused, searchMatchCountByTopology,
      isResourceViewMode, searchQuery, topologiesLoaded, inputId = 'search'
    } = this.props;
    const hidden = !topologiesLoaded || isResourceViewMode;
    const disabled = this.props.isTopologyNodeCountZero && !hidden;
    const matchCount = searchMatchCountByTopology
      .reduce((count, topologyMatchCount) => count + topologyMatchCount, 0);
    const showPinnedSearches = pinnedSearches.size > 0;
    // manual clear (null) has priority, then props, then state
    const value = this.state.value === null ? '' : this.state.value || searchQuery || '';
    const classNames = classnames('search', 'hideable', {
      hide: hidden,
      'search-pinned': showPinnedSearches,
      'search-matched': matchCount,
      'search-filled': value,
      'search-focused': searchFocused,
      'search-disabled': disabled
    });
    const title = matchCount ? `${matchCount} matches` : null;

    return (
      <div className="search-wrapper">
        <div className={classNames} title={title}>
          <div className="search-input">
            {showPinnedSearches && pinnedSearches.toIndexedSeq()
              .map(query => <SearchItem query={query} key={query} />)}
            <input
              className="search-input-field" type="text" id={inputId}
              value={value} onChange={this.handleChange} onKeyUp={this.handleKeyUp}
              onFocus={this.handleFocus} onBlur={this.handleBlur}
              disabled={disabled} ref={this.saveQueryInputRef} />
          </div>
          <div className="search-label">
            <i className="fa fa-search search-label-icon" />
            <span className="search-label-hint" htmlFor={inputId}>
              Search
            </span>
          </div>
          {!showPinnedSearches && <div className="search-hint">
            {getHint(nodes)} <span
              className="search-help-link fa fa-question-circle"
              onMouseDown={this.props.toggleHelp} />
          </div>}
        </div>
      </div>
    );
  }
}


export default connect(
  state => ({
    nodes: state.get('nodes'),
    topologyViewMode: state.get('topologyViewMode'),
    isResourceViewMode: isResourceViewModeSelector(state),
    isTopologyNodeCountZero: isTopologyNodeCountZero(state),
    currentTopology: state.get('currentTopology'),
    topologiesLoaded: state.get('topologiesLoaded'),
    pinnedSearches: state.get('pinnedSearches'),
    searchFocused: state.get('searchFocused'),
    searchQuery: state.get('searchQuery'),
    searchMatchCountByTopology: searchMatchCountByTopologySelector(state),
  }),
  {
    blurSearch, doSearch, focusSearch, pinSearch, toggleHelp
  }
)(Search);
