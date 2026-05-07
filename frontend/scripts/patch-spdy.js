#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

const files = [
  'node_modules/spdy/lib/spdy/server.js',
  'node_modules/spdy/lib/spdy/agent.js',
  'node_modules/spdy-transport/lib/spdy-transport/utils.js',
  'node_modules/spdy/test/client-test.js'
];

const replacement = `function(target, source) {
      for (var key in source) {
        if (source.hasOwnProperty(key)) {
          target[key] = source[key]
        }
      }
      return target
    }`;

files.forEach(file => {
  const filePath = path.join(__dirname, '..', file);
  
  if (!fs.existsSync(filePath)) {
    return;
  }
  
  let content = fs.readFileSync(filePath, 'utf8');
  
  if (content.includes('util._extend')) {
    content = content.replace(/util\._extend/g, replacement);
    fs.writeFileSync(filePath, content, 'utf8');
    console.log(`✓ Patched ${file}`);
  }
});
