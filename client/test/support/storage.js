const makeStorageMock = function () {
  let store = {};
  return {
    store,
    getItem(key) {
      return store[key];
    },
    setItem(key, value) {
      store[key] = value;
    },
    clear() {
      store = {};
    }
  };
};

const localStorageMock = makeStorageMock();
const sessionStorageMock = makeStorageMock();

Object.defineProperty(window, 'localStorage', { value: localStorageMock });
Object.defineProperty(window, 'sessionStorage', { value: sessionStorageMock });
