import React from 'react';
import { connect } from 'react-redux';

/**
 * Returns an array with chunks that cover the whole text via {start, length}
 * objects.
 *
 * `([{start: 2, length: 1}], "text") =>
 *   [{start: 0, length: 2}, {start: 2, length: 1, match: true}, {start: 3, length: 1}]`
 */
function reduceMatchesToChunks(matches, text) {
  if (text && matches && matches.length > 0) {
    const result = matches.reduce((chunks, match) => {
      const prev = chunks.length > 0 ? chunks[chunks.length - 1] : null;
      const end = prev ? prev.start + prev.length : 0;
      // skip non-matching chunk if first chunk is match
      if (match.start > 0) {
        chunks.push({start: end, length: match.start});
      }
      chunks.push(Object.assign({match: true}, match));
      return chunks;
    }, []);
    const last = result[result.length - 1];
    const remaining = last.start + last.length;
    if (text && remaining < text.length) {
      result.push({start: remaining, length: text.length - remaining});
    }
    return result;
  }
  return [];
}

/**
 * Renders text with highlighted search matches.
 *
 * `props.matches` must be an immutable.Map of match
 * objects, the match object for this component will be extracted
 * via `get(props.fieldId)`).
 * A match object is of shape `{text, label, matches}`.
 * `match.matches` is an array of text matches of shape `{start, length}`
 * that delimit text matches in `text`. `label` shows the origin of the text.
 */
class MatchedText extends React.Component {

  render() {
    const { fieldId, matches, text } = this.props;
    // match is a direct match object, or still need to extract the correct field
    const fieldMatches = matches && matches.get(fieldId);

    if (!fieldMatches) {
      return <span>{text}</span>;
    }

    return (
      <span className="matched-text">
        {reduceMatchesToChunks(fieldMatches.matches, text).map((chunk, index) => {
          if (chunk.match) {
            return (
              <span className="match" key={index} title={`matched: ${fieldMatches.label}`}>
                {text.substr(chunk.start, chunk.length)}
              </span>
            );
          }
          return text.substr(chunk.start, chunk.length);
        })}
      </span>
    );
  }
}

export default connect()(MatchedText);
