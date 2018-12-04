import debug from 'debug';

const log = debug('service:tracking');

// Track segment events only if Scope is running inside of Weave Cloud.
export function trackAnalyticsEvent(name, props) {
  if (window.analytics && process.env.WEAVE_CLOUD) {
    window.analytics.track(name, props);
  } else {
    log('trackAnalyticsEvent', name, props);
  }
}
