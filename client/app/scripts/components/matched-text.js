import React from 'react';

const TRUNCATE_CONTEXT = 6;
const TRUNCATE_ELLIPSIS = '…';

/**
 * Returns an array with chunks that cover the whole text via {start, length}
 * objects.
 *
 * `('text', {start: 2, length: 1}) => [{text: 'te'}, {text: 'x', match: true}, {text: 't'}]`
 */
function chunkText(text, { start, length }) {
  if (text && !window.isNaN(start) && !window.isNaN(length)) {
    const chunks = [];
    // text chunk before match
    if (start > 0) {
      chunks.push({text: text.substr(0, start)});
    }
    // matching chunk
    chunks.push({match: true, offset: start, text: text.substr(start, length)});
    // text after match
    const remaining = start + length;
    if (remaining < text.length) {
      chunks.push({text: text.substr(remaining)});
    }
    return chunks;
  }
  return [{ text }];
}

/**
 * Truncates chunks with ellipsis
 *
 * First chunk is truncated from left, second chunk (match) is truncated in the
 * middle, last chunk is truncated at the end, e.g.
 * `[{text: "...cation is a "}, {text: "useful...or not"}, {text: "tool..."}]`
 */
function truncateChunks(chunks, text, maxLength) {
  if (chunks && chunks.length === 3 && maxLength && text && text.length > maxLength) {
    const res = chunks.map(c => Object.assign({}, c));
    let needToCut = text.length - maxLength;
    // trucate end
    const end = res[2];
    if (end.text.length > TRUNCATE_CONTEXT) {
      needToCut -= end.text.length - TRUNCATE_CONTEXT;
      end.text = `${end.text.substr(0, TRUNCATE_CONTEXT)}${TRUNCATE_ELLIPSIS}`;
    }

    if (needToCut) {
      // truncate front
      const start = res[0];
      if (start.text.length > TRUNCATE_CONTEXT) {
        needToCut -= start.text.length - TRUNCATE_CONTEXT;
        start.text = `${TRUNCATE_ELLIPSIS}`
          + `${start.text.substr(start.text.length - TRUNCATE_CONTEXT)}`;
      }
    }

    if (needToCut) {
      // truncate match
      const middle = res[1];
      if (middle.text.length > 2 * TRUNCATE_CONTEXT) {
        middle.text = `${middle.text.substr(0, TRUNCATE_CONTEXT)}`
          + `${TRUNCATE_ELLIPSIS}`
          + `${middle.text.substr(middle.text.length - TRUNCATE_CONTEXT)}`;
      }
    }

    return res;
  }
  return chunks;
}

/**
 * Renders text with highlighted search match.
 *
 * A match object is of shape `{text, label, match}`.
 * `match` is a text match object of shape `{start, length}`
 * that delimit text matches in `text`. `label` shows the origin of the text.
 */
export default class MatchedText extends React.PureComponent {
  render() {
    const {
      match, text, truncate, maxLength
    } = this.props;

    const showFullValue = !truncate || (match && (match.start + match.length) > truncate);
    const displayText = showFullValue ? text : text.slice(0, truncate);

    if (!match) {
      return <span>{displayText}</span>;
    }

    const chunks = chunkText(displayText, match);

    return (
      <span className="matched-text" title={text}>
        {truncateChunks(chunks, displayText, maxLength).map((chunk) => {
          if (chunk.match) {
            return (
              <span className="match" key={chunk.offset}>
                {chunk.text}
              </span>
            );
          }
          return chunk.text;
        })}
      </span>
    );
  }
}
