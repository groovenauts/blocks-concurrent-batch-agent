const path = require('path');

const distFiles = ['favicon.ico', 'main.bundle.js'];

module.exports = {
  entry: {
    main: './src/index.js',
    test: './src/test.js',
  },
  output: {
    filename: '[name].bundle.js',
    path: path.resolve(__dirname, 'dist')
  },
  module: {
    rules: [
      { test: /\.js$/, exclude: /node_modules/, loader: "babel-loader" },
      {
        test: /\.css$/,
        use: [
          'style-loader',
          'css-loader'
        ]
      },
      // // node_modules/ を含む .svg ファイルにヒットする
      { test: /node_modules\/.+\.svg$/, loader: "svg-url-loader" },
      // // 否定先読みで node_modules/ を含まない .svg ファイルにヒットする
      { test: /^(?!.*node_modules\/).+\.svg$/, loader: "react-svg-loader" }

    ]
  },
  devServer: {
    contentBase: path.resolve('dist'),
    publicPath: '/',
    before: function(app){
      app.get('/*', function(req, res) {
        reqFile = req.path.slice(1)
        file = distFiles.includes(reqFile) ? reqFile : "index.html"
        res.sendFile(path.resolve(__dirname, 'dist/' + file));
      });
    }
  }
};
