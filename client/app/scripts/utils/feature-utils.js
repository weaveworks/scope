import { storageGet, storageSet } from './storage-utils';

// prefix for all feature flags
const STORAGE_KEY_PREFIX = 'scope-experimental:';

const getKey = key => `${STORAGE_KEY_PREFIX}${key}`;

/**
 * Returns true if `feature` is enabled
 *
 * Features can be enabled either via calling `setFeature()` or by setting
 * `localStorage.scope-experimental:featureName = true` in the console.
 * @param  {String} feature Feature name, ideally one word or hyphenated
 * @return {Boolean}         True if feature is enabled
 */
export function featureIsEnabled(feature) {
  let enabled = storageGet(getKey(feature));
  if (typeof enabled === 'string') {
    // Convert back to boolean if stored as a string.
    enabled = JSON.parse(enabled);
  }
  return enabled;
}

/**
 * Returns true if any of the features given as arguments are enabled.
 *
 * Useful if features are hierarchical, e.g.:
 * `featureIsEnabledAny('superFeature', 'subFeature')`
 * @param  {String} args Feature names
 * @return {Boolean}      True if any of the features are enabled
 */
export function featureIsEnabledAny(...args) {
  return Array.prototype.some.call(args, feature => featureIsEnabled(feature));
}

/**
 * Set true/false if a feature is enabled.
 * @param {String}  feature   Feature name
 * @param {Boolean} isEnabled true/false
 */
export function setFeature(feature, isEnabled) {
  return storageSet(getKey(feature), isEnabled);
}
