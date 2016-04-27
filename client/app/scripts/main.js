if (process.env.NODE_ENV === 'production') {
  module.exports = require('./main.prod');
} else {
  module.exports = require('./main.dev');
}
