
const PREFIX = document.title || 'Weave Scope';
const SEPARATOR = ' - ';

export function setDocumentTitle(title) {
  if (title) {
    document.title = [PREFIX, title].join(SEPARATOR);
  } else {
    document.title = PREFIX;
  }
}

export function resetDocumentTitle() {
  setDocumentTitle(null);
}
