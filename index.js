const express = require('express');
const { createProxyMiddleware, fixRequestBody } = require('http-proxy-middleware');
const dotenv = require('dotenv');
dotenv.config();

const app = express();

let patsArray = process.env.PATS.split(',');

const mapping = {}

app.use(express.json())

app.use('/', createProxyMiddleware({
	target: 'https://api.github.com/graphql',
	pathRewrite: { '^/': '' },
	changeOrigin: true,
	onProxyReq: (proxyReq, req, res) => {
		let token = patsArray[Math.floor(Math.random() * patsArray.length)];
		proxyReq.setHeader('Authorization', `Bearer ${token}`);
	
		// hash the query and variables to get a unique key
		const key = `${JSON.stringify(req.body.query)}${JSON.stringify(req.body.variables)}`
		
		if (mapping[key]) {
			return mapping[key]
		} else {
			mapping[key] = 'filler'
		}
		
		// this method provided by http-proxy-middleware fixes the body after bodyParser has it's way with it
		fixRequestBody(proxyReq, req)
	}
}));

app.listen(3000, () => {
	console.log('Server started on port 3000')
})

// app.listen with start message