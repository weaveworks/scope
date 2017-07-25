import debug from 'debug';

const log = debug('scope:tracking');

// Track mixpanel events only if Scope is running inside of Weave Cloud.
export function trackMixpanelEvent(name, props) {
  log('trackMixpanelEvent', name, props);
  if (window.mixpanel && process.env.WEAVE_CLOUD) {
    window.mixpanel.track(name, props);
  } else {
    log('trackMixpanelEvent', name, props);
  }
}
