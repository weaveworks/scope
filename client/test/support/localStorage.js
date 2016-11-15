const localStorageMock = (function() {
  let store = {};
  return {
    store,
    getItem: function(key) {
      return store[key];
    },
    setItem: function(key, value) {
      store[key] = value;
    },
    clear: function() {
      store = {};
    }
  };
})();
Object.defineProperty(window, 'Storage', { value: localStorageMock });
Object.defineProperty(window, 'localStorage', { value: localStorageMock });
