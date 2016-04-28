import React from 'react';
import { connect } from 'react-redux';

import MatchedText from './matched-text';

const SHOW_ROW_COUNT = 3;

class MatchedResults extends React.Component {

  renderMatch(matches, field) {
    const match = matches.get(field);
    return (
      <div className="matched-results-match" key={match.label}>
        <div className="matched-results-match-wrapper">
          <span className="matched-results-match-label">
            {match.label}:
          </span>
          <MatchedText text={match.text} matches={matches} fieldId={field} />
        </div>
      </div>
    );
  }

  render() {
    const { matches } = this.props;

    if (!matches) {
      return null;
    }

    let moreFieldMatches;
    let moreFieldMatchesTitle;
    if (matches.size > SHOW_ROW_COUNT) {
      moreFieldMatches = matches
        .valueSeq()
        .skip(SHOW_ROW_COUNT)
        .map(field => field.label);
      moreFieldMatchesTitle = `More matches:\n${moreFieldMatches.join(',\n')}`;
    }

    return (
      <div className="matched-results">
        {matches.keySeq().take(SHOW_ROW_COUNT).map(fieldId => this.renderMatch(matches, fieldId))}
        {moreFieldMatches && <span className="matched-results-more" title={moreFieldMatchesTitle}>
          {`${moreFieldMatches.size} more matches`}
        </span>}
      </div>
    );
  }
}

export default connect()(MatchedResults);
