import { storageGetObject, storageSetObject } from './storage-utils';

const STORAGE_FEATURE_KEY = 'scopeFeatureFlags';

export function featureIsEnabled(feature) {
  const features = storageGetObject(STORAGE_FEATURE_KEY, {});
  return features[feature];
}

export function setFeature(feature, isEnabled) {
  const features = storageGetObject(STORAGE_FEATURE_KEY, {});
  features[feature] = isEnabled;
  return storageSetObject(STORAGE_FEATURE_KEY, features);
}

export function clearFeatures() {
  return storageSetObject(STORAGE_FEATURE_KEY, {});
}
