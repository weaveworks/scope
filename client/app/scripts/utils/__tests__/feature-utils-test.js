
const FU = require('../feature-utils');

describe('FeatureUtils', () => {
  const FEATURE_X_KEY = 'my feature 1';
  const FEATURE_Y_KEY = 'my feature 2';

  beforeEach(() => {
    FU.setFeature(FEATURE_X_KEY, false);
    FU.setFeature(FEATURE_Y_KEY, false);
  });

  describe('Setting of features', () => {
    it('should not have any features by default', () => {
      expect(FU.featureIsEnabled(FEATURE_X_KEY)).toBeFalsy();
      expect(FU.featureIsEnabled(FEATURE_Y_KEY)).toBeFalsy();
    });

    it('should work with enabling one feature', () => {
      let success;
      expect(FU.featureIsEnabled(FEATURE_X_KEY)).toBeFalsy();
      success = FU.setFeature(FEATURE_X_KEY, true);
      expect(success).toBeTruthy();
      expect(FU.featureIsEnabled(FEATURE_X_KEY)).toBeTruthy();
      expect(FU.featureIsEnabled(FEATURE_Y_KEY)).toBeFalsy();
      success = FU.setFeature(FEATURE_X_KEY, false);
      expect(success).toBeTruthy();
      expect(FU.featureIsEnabled(FEATURE_X_KEY)).toBeFalsy();
    });

    it('should allow for either feature', () => {
      let success;
      expect(FU.featureIsEnabledAny(FEATURE_X_KEY, FEATURE_Y_KEY)).toBeFalsy();
      success = FU.setFeature(FEATURE_X_KEY, true);
      expect(success).toBeTruthy();
      expect(FU.featureIsEnabledAny(FEATURE_X_KEY, FEATURE_Y_KEY)).toBeTruthy();
      success = FU.setFeature(FEATURE_X_KEY, false);
      success = FU.setFeature(FEATURE_Y_KEY, true);
      expect(success).toBeTruthy();
      expect(FU.featureIsEnabledAny(FEATURE_X_KEY, FEATURE_Y_KEY)).toBeTruthy();
    });
  });
});
