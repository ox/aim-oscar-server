export function char(num : number) : Buffer {
  const buf = Buffer.alloc(1, 0x00);
  buf.writeUInt8(num);
  return buf;
}

export function word(num : number) : Buffer {
  const buf = Buffer.alloc(2, 0x00);
  buf.writeUInt16BE(num);
  return buf;
}

export function dword(num : number) : Buffer {
  const buf = Buffer.alloc(4, 0x00);
  buf.writeUInt32BE(num);
  return buf;
}

export function qword(num : number) : Buffer {
  const buf = Buffer.alloc(8, 0x00);
  buf.writeUInt32BE(num);
  return buf;
}

/**
 * Converts a string IP address to it's number representation.
 * From: https://stackoverflow.com/a/8105740
 * @param ip IP address string
 * @returns IP address as a number
 */
export function dot2num(ip : string) : number {
    const d = ip.split('.');
    return ((((((+d[0])*256)+(+d[1]))*256)+(+d[2]))*256)+(+d[3]);
}

export function num2dot(num : number) : string {
    let d = '' + num%256;
    for (var i = 3; i > 0; i--) { 
        num = Math.floor(num/256);
        d = num%256 + '.' + d;
    }
    return d;
}
