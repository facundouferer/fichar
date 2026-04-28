const { createServer } = require('http');
const { join, dirname } = require('path');
const { existsSync, createReadStream, statSync } = require('fs');
const path = require('path');

const currentDir = process.cwd();
const buildDir = join(currentDir, 'dist');

const MIME_TYPES = {
  '.html': 'text/html',
  '.js': 'application/javascript',
  '.css': 'text/css',
  '.json': 'application/json',
  '.png': 'image/png',
  '.jpg': 'image/jpeg',
  '.svg': 'image/svg+xml',
  '.ico': 'image/x-icon',
};

const PORT = process.env.PORT || 4321;
const BACKEND_URL = process.env.BACKEND_URL || 'http://backend:8082';
const ALLOWED_HOSTS = ['localhost', '127.0.0.1', 'fichar.gar.com.ar', 'www.fichar.gar.com.ar'];

const server = createServer(async (req, res) => {
  const host = req.headers.host;
  const allowed = ALLOWED_HOSTS.some(h => host === h || host === `${h}:${PORT}`);
  if (!allowed) {
    res.writeHead(403, { 'Content-Type': 'text/plain' });
    res.end(`Blocked request. This host ("${host}") is not allowed.`);
    return;
  }

  let url = req.url || '/';
  
  if (url.startsWith('/api/')) {
    try {
      const body = await new Promise((resolve, reject) => {
        let data = '';
        req.on('data', chunk => data += chunk);
        req.on('end', () => resolve(data));
        req.on('error', reject);
      });

      const response = await fetch(`${BACKEND_URL}${url}`, {
        method: req.method,
        headers: Object.fromEntries(
          Object.entries(req.headers).filter(([k]) => !['host', 'connection', 'content-length'].includes(k))
        ),
        body: body || undefined,
      });
      
      const data = await response.text();
      res.writeHead(response.status, { 'Content-Type': 'application/json' });
      res.end(data);
    } catch (err) {
      res.writeHead(502, { 'Content-Type': 'text/plain' });
      res.end('Backend error: ' + err.message);
    }
    return;
  }
  
  url = url.split('?')[0];
  
  let filePath = join(buildDir, url.slice(1));
  
  let isDir = false;
  try {
    isDir = statSync(filePath).isDirectory();
  } catch (e) {}
  
  if (!existsSync(filePath) || isDir) {
    filePath = join(filePath, 'index.html');
  }
  
  if (!existsSync(filePath)) {
    filePath = join(buildDir, 'index.html');
  }
  
  if (existsSync(filePath)) {
    const ext = path.extname(filePath) || '.html';
    const contentType = MIME_TYPES[ext] || 'text/plain';
    
    res.writeHead(200, { 'Content-Type': contentType });
    createReadStream(filePath).pipe(res);
  } else {
    res.writeHead(404, { 'Content-Type': 'text/plain' });
    res.end('Not Found');
  }
});

server.listen(PORT, '0.0.0.0', () => {
  console.log(`Frontend server running on http://0.0.0.0:${PORT}`);
  console.log(`API proxy to ${BACKEND_URL}`);
});