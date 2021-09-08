function chunkString(str, len) {
  const size = Math.ceil(str.length/len)
  const r = Array(size)
  let offset = 0
  
  for (let i = 0; i < size; i++) {
    r[i] = str.substr(offset, len)
    offset += len
  }
  
  return r
}

function logDataStream(data){  
  const strs = chunkString(data.toString('hex'), 16);
  return strs.map((str) => chunkString(str, 2).join(' ')).join('\n');
}

module.exports = {
  logDataStream,
};
