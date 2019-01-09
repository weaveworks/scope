import React from 'react';
import { connect } from 'react-redux';
import { isEmpty } from 'lodash';
import { Search } from 'weaveworks-ui-components';
import styled from 'styled-components';

import { blurSearch, focusSearch, updateSearch, toggleHelp } from '../actions/app-actions';
import { searchMatchCountByTopologySelector } from '../selectors/search';
import { isResourceViewModeSelector } from '../selectors/topology';
import { slugify } from '../utils/string-utils';
import { isTopologyNodeCountZero } from '../utils/topology-utils';
import { trackAnalyticsEvent } from '../utils/tracking-utils';


const SearchWrapper = styled.div`
  margin: 0 8px;
  min-width: 160px;
  text-align: right;
`;

const SearchContainer = styled.div`
  display: inline-block;
  position: relative;
  pointer-events: all;
  line-height: 100%;
  max-width: 400px;
  width: 100%;
`;

const SearchHint = styled.div`
  font-size: ${props => props.theme.fontSizes.tiny};
  color: ${props => props.theme.colors.purple400};
  transition: transform 0.3s 0s ease-in-out, opacity 0.3s 0s ease-in-out;
  text-align: left;
  margin-top: 3px;
  padding: 0 1em;
  opacity: 0;

  ${props => props.active && `
    opacity: 1;
  `};
`;

const SearchHintIcon = styled.span`
  font-size: ${props => props.theme.fontSizes.normal};
  cursor: pointer;

  &:hover {
    color: ${props => props.theme.colors.purple600};
  }
`;

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
    [label] = shortenHintLabel(node.get('label')).split('.');
    if (node.get('metadata')) {
      const metadataField = node.get('metadata').first();
      metadataLabel = shortenHintLabel(slugify(metadataField.get('label')))
        .split('.').pop();
      metadataValue = shortenHintLabel(metadataField.get('value'));
    }
  }

  return `Try "${label}", "${metadataLabel}:${metadataValue}", or "cpu > 2%".`;
}


class SearchComponent extends React.Component {
  handleChange = (searchQuery, pinnedSearches) => {
    trackAnalyticsEvent('scope.search.query.change', {
      layout: this.props.topologyViewMode,
      parentTopologyId: this.props.currentTopology.get('parentId'),
      topologyId: this.props.currentTopology.get('id'),
    });
    this.props.updateSearch(searchQuery, pinnedSearches);
  }

  render() {
    const {
      searchHint, searchMatchesCount, searchQuery, pinnedSearches, topologiesLoaded,
      isResourceViewMode, isTopologyEmpty,
    } = this.props;

    return (
      <SearchWrapper>
        <SearchContainer title={searchMatchesCount ? `${searchMatchesCount} matches` : undefined}>
          <Search
            placeholder="search"
            query={searchQuery}
            pinnedTerms={pinnedSearches}
            disabled={topologiesLoaded && !isResourceViewMode && isTopologyEmpty}
            onChange={this.handleChange}
            onFocus={this.props.focusSearch}
            onBlur={this.props.blurSearch}
          />
          <SearchHint active={this.props.searchFocused && isEmpty(pinnedSearches)}>
            {searchHint} <SearchHintIcon
              className="fa fa-question-circle"
              onMouseDown={this.props.toggleHelp}
            />
          </SearchHint>
        </SearchContainer>
      </SearchWrapper>
    );
  }
}


export default connect(
  state => ({
    currentTopology: state.get('currentTopology'),
    isResourceViewMode: isResourceViewModeSelector(state),
    isTopologyEmpty: isTopologyNodeCountZero(state),
    pinnedSearches: state.get('pinnedSearches').toJS(),
    searchFocused: state.get('searchFocused'),
    searchHint: getHint(state.get('nodes')),
    searchMatchesCount: searchMatchCountByTopologySelector(state)
      .reduce((count, topologyMatchCount) => count + topologyMatchCount, 0),
    searchQuery: state.get('searchQuery'),
    topologiesLoaded: state.get('topologiesLoaded'),
    topologyViewMode: state.get('topologyViewMode'),
  }),
  {
    blurSearch, focusSearch, toggleHelp, updateSearch
  }
)(SearchComponent);
