const path = require('path')
// var HtmlWebpackPlugin = require('html-webpack-plugin')

module.exports = {
  mode: 'development',
  entry: './src/index.ts',
  devtool: 'inline-source-map',
  // plugins: [
  //   new HtmlWebpackPlugin({
  //     template: './src/template.html',
  //   }),
  // ],
  module: {
    rules: [
      // Handle TypeScript
      {
        test: /\.tsx?$/,
        use: 'ts-loader',
        exclude: [/node_modules/],
      },
      // Handle our workers
      {
        test: /\.worker\.js$/,
        use: { loader: 'worker-loader' },
      },
    ],
  },
  resolve: {
    extensions: ['.ts', '.js'],
  },
  output: {
    // This is required so workers are known where to be loaded from
    publicPath: '/dist/',
    filename: 'bundle.js',
    path: path.resolve(__dirname, 'dist/'),
  },
}
