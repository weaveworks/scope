
const STANDALONE_TITLE = 'Weave Scope';
const STANDALONE = document.title === STANDALONE_TITLE;
const SEPARATOR = ' – ';

export function setDocumentTitle(title) {
  if (!STANDALONE) {
    return;
  }
  if (title) {
    document.title = [STANDALONE_TITLE, title].join(SEPARATOR);
  } else {
    document.title = STANDALONE_TITLE;
  }
}

export function resetDocumentTitle() {
  setDocumentTitle(null);
}
