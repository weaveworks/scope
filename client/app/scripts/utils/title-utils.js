
const PREFIX = 'Container';

export function setDocumentTitle(title) {
  if (title) {
    document.title = title;
  } else {
    document.title = PREFIX;
  }
}

export function resetDocumentTitle() {
  setDocumentTitle(null);
}
