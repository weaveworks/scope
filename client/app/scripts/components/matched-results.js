import React from 'react';
import { MatchedText } from 'weaveworks-ui-components';

const SHOW_ROW_COUNT = 2;
const MAX_MATCH_LENGTH = 24;


const Match = match => (
  <div className="matched-results-match" key={match.label}>
    <div className="matched-results-match-wrapper">
      <span className="matched-results-match-label">
        {match.label}:
      </span>
      <MatchedText
        noBorder
        text={match.text}
        match={match}
        maxLength={MAX_MATCH_LENGTH}
        truncate={match.truncate}
      />
    </div>
  </div>
);

export default class MatchedResults extends React.PureComponent {
  render() {
    const { matches, style } = this.props;

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
      <div className="matched-results" style={style}>
        {matches.keySeq().take(SHOW_ROW_COUNT).map(fieldId => Match(matches.get(fieldId)))}
        {moreFieldMatches &&
          <div className="matched-results-more" title={moreFieldMatchesTitle}>
            {`${moreFieldMatches.size} more matches`}
          </div>
        }
      </div>
    );
  }
}
