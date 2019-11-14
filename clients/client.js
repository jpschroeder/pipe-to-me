const http = require('https');
if (process.argv.length < 3 || process.argv[2] === '-h' || process.argv[2] === '-h') {
	console.log('usage: node client.js https://pipeto.me/<code>');
	return;
}

console.log('connected to: ' + process.argv[2]);
var req = http.request(process.argv[2], {
    method: 'PUT',
    headers: { 'Expect': '100-continue' }
}, resp => resp.pipe(process.stdout));
process.stdin.pipe(req);