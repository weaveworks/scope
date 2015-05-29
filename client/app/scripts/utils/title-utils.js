
const PREFIX = 'Weave Scope';
const SEPARATOR = ' - ';

function setDocumentTitle(title) {
  if (title) {
    document.title = [PREFIX, title].join(SEPARATOR);
  } else {
    document.title = PREFIX;
  }
}

function resetDocumentTitle() {
  setDocumentTitle(null);
}

module.exports = {
  resetTitle: resetDocumentTitle,
  setTitle: setDocumentTitle
};
