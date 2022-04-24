const express = require('express');
const app = express();
var httpProxy = require('http-proxy');
const port = 3000;
var apiProxy = httpProxy.createProxyServer();
var githubApi = 'https://api.github.com/graphql';

app.get("/ghproxy", (req, res) => {
	console.log('Redirecting to GitHub API...');
	apiProxy.web(req, res, { target: githubApi, changeOrigin: true });
});

app.listen(port, () => {
	console.log(`App listening on port ${port}`);
});