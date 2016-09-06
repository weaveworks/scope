import debug from 'debug';

const log = debug('scope:storage-utils');

// localStorage detection
const storage = typeof(Storage) !== 'undefined' ? window.localStorage : null;

export function storageGet(key, defaultValue) {
  if (storage && storage[key] !== undefined) {
    return storage.getItem(key);
  }
  return defaultValue;
}

export function storageSet(key, value) {
  if (storage) {
    try {
      storage.setItem(key, value);
    } catch (e) {
      log('Error storing value in storage. Maybe full? Could not store key.', key);
    }
  }
}
