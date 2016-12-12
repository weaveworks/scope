import debug from 'debug';

const log = debug('scope:storage-utils');

// localStorage detection
const storage = typeof(Storage) !== 'undefined' ? window.localStorage : null;

export function storageGet(key, defaultValue) {
  if (storage && storage.getItem(key) !== undefined) {
    return storage.getItem(key);
  }
  return defaultValue;
}

export function storageSet(key, value) {
  if (storage) {
    try {
      storage.setItem(key, value);
      return true;
    } catch (e) {
      log('Error storing value in storage. Maybe full? Could not store key.', key);
    }
  }
  return false;
}

export function storageGetObject(key, defaultValue) {
  const value = storageGet(key);
  if (value) {
    try {
      return JSON.parse(value);
    } catch (e) {
      log('Error getting object for key.', key);
    }
  }
  return defaultValue;
}

export function storageSetObject(key, obj) {
  try {
    return storageSet(key, JSON.stringify(obj));
  } catch (e) {
    log('Error encoding object for key', key);
  }
  return false;
}
