import { createServer } from 'node:http';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { readFileSync, existsSync } from 'node:fs';
import { createReadStream } from 'node:fs';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const buildDir = join(__dirname, '../dist');
const indexHtmlPath = join(buildDir, 'index.html');

let indexHtml = '';
if (existsSync(indexHtmlPath)) {
  indexHtml = readFileSync(indexHtmlPath, 'utf-8');
}

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

const server = createServer((req, res) => {
  let url = req.url || '/';
  
  // Remove query string
  url = url.split('?')[0];
  
  // Default to index.html for SPA routing
  let filePath = join(buildDir, url === '/' ? 'index.html' : url);
  
  // Check if file exists, otherwise serve index.html for SPA
  if (!existsSync(filePath)) {
    filePath = join(buildDir, 'index.html');
  }
  
  if (existsSync(filePath)) {
    const ext = join(filePath, '.').split('.').pop() || '.html';
    const contentType = MIME_TYPES[ext] || 'text/plain';
    
    res.writeHead(200, { 'Content-Type': contentType });
    createReadStream(filePath).pipe(res);
  } else {
    res.writeHead(404, { 'Content-Type': 'text/plain' });
    res.end('Not Found');
  }
});

const PORT = process.env.PORT || 4321;
server.listen(PORT, () => {
  console.log(`Frontend server running on http://localhost:${PORT}`);
});