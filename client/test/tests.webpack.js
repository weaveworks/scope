var context = require.context('../app/scripts', true, /-test\.js$/);
context.keys().forEach(context);
