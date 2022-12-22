const express = require('express');
const { createProxyMiddleware, fixRequestBody, responseInterceptor } = require('http-proxy-middleware');
const dotenv = require('dotenv');
dotenv.config();

const app = express();

let patsArray = process.env.PATS.split(',');

const mapping = {}

app.use(express.json())

app.use('/', createProxyMiddleware({
	target: 'https://api.github.com/graphql',
	pathRewrite: { '^/': '' },
	selfHandleResponse: true,
	changeOrigin: true,
	onProxyReq: (proxyReq, req, res) => {
		let token = patsArray[Math.floor(Math.random() * patsArray.length)];
		proxyReq.setHeader('Authorization', `Bearer ${token}`);
	
		// hash the query and variables to get a unique key
		const key = `${JSON.stringify(req.body.query)}${JSON.stringify(req.body.variables)}`
		
		if (mapping[key]) {
			return res.json(JSON.parse(mapping[key]))
		} else {
			mapping[key] = 'filler'
		}
		
		// this method provided by http-proxy-middleware fixes the body after bodyParser has it's way with it
		fixRequestBody(proxyReq, req)
	},
	onProxyRes: responseInterceptor(async (responseBuffer, proxyRes, req, res) => {
		const key = `${JSON.stringify(req.body.query)}${JSON.stringify(req.body.variables)}`
    const response = responseBuffer.toString('utf8');
    mapping[key] = response
		return response
  })
}));

app.listen(3000, () => {
	console.log('Server started on port 3000')
})

// app.listen with start message