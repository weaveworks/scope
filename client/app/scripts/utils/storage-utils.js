import debug from 'debug';

const log = debug('scope:storage-utils');

export const localSessionStorage = {
  clear() {
    window.sessionStorage.clear();
    window.localStorage.clear();
  },
  getItem(k) {
    return window.sessionStorage.getItem(k) || window.localStorage.getItem(k);
  },
  setItem(k, v) {
    window.sessionStorage.setItem(k, v);
    window.localStorage.setItem(k, v);
  }
};

export function storageGet(key, defaultValue, storage = localSessionStorage) {
  if (!storage) {
    return defaultValue;
  }

  const value = storage.getItem(key);
  if (value == null) {
    return defaultValue;
  }

  return value;
}

export function storageSet(key, value, storage = localSessionStorage) {
  if (storage) {
    try {
      storage.setItem(key, value);
      return true;
    } catch (e) {
      log(
        'Error storing value in storage. Maybe full? Could not store key.',
        key
      );
    }
  }
  return false;
}

export function storageGetObject(
  key,
  defaultValue,
  storage = localSessionStorage
) {
  const value = storageGet(key, undefined, storage);
  if (value) {
    try {
      return JSON.parse(value);
    } catch (e) {
      log('Error getting object for key.', key);
    }
  }
  return defaultValue;
}

export function storageSetObject(key, obj, storage = localSessionStorage) {
  try {
    return storageSet(key, JSON.stringify(obj), storage);
  } catch (e) {
    log('Error encoding object for key', key);
  }
  return false;
}
