import { FLAP } from './structures'
import { dot2num, num2dot } from './structures/bytes'
export function chunkString(str : string, len : number) {
  const size = Math.ceil(str.length/len)
  const r = Array(size)
  let offset = 0
  
  for (let i = 0; i < size; i++) {
    r[i] = str.substr(offset, len)
    offset += len
  }
  
  return r
}

export function logDataStream(data : Buffer){  
  const strs = chunkString(data.toString('hex'), 16);
  return strs.map((str) => chunkString(str, 2).join(' ')).join('\n');
}

// Experiment to provide descriptive print-outs of the data structures
// @ts-ignore
import Table from 'table-layout';

type DataSize = -1 | 8 | 16 | 32;
interface Spec {
  description: string,
  size : DataSize,
  isRepeat? : boolean,
  isParam? : boolean,
  isTLV? : boolean,
  repeatSpecs?: Spec[],
}

function byte(description : string) : Spec {
  return {size: 8, description}
}

function word(description : string) : Spec {
  return {size: 16, description};
}

function dword(description : string) : Spec {
  return {size: 32, description};
}

function repeat(size: DataSize, description : string, specs : Spec[]) : Spec {
  return {size, description, isRepeat: true, repeatSpecs: specs};
}

function param(size: DataSize, description: string) : Spec {
  return {size, description, isParam: true};
}

function tlv(description : string) : Spec {
  return {size : -1, description, isTLV: true};
}

function parseBuffer(buf : Buffer, spec : Spec[], repeatTimes = 0) {
  let offset = 0;
  let rows = [];
  let repeat = repeatTimes;

  for (let section of spec) {
    let value : any = 0;
    let bufStr : string = '';
    if (section.size === 8) {
      bufStr = buf.slice(offset, offset + 1).toString('hex');
      value = buf.readInt8(offset);
      offset += 1;
    } else if (section.size === 16) {
      bufStr = buf.slice(offset, offset + 2).toString('hex');
      value = buf.readUInt16BE(offset);
      offset += 2;
    } else if (section.size === 32) {
      bufStr = buf.slice(offset, offset + 4).toString('hex');
      value = buf.readUInt32BE(offset);
      offset += 4;
    }

    if (section.description.includes("IP")) {
      value = num2dot(value);
    }


    if (section.isParam) {
      const paramBuf = buf.slice(offset, offset + value);
      offset += value;
      rows.push([chunkString(paramBuf.toString('hex'), 2), paramBuf.toString('ascii'), section.description])
    } else if (section.isTLV) {
      const tlvType = buf.slice(offset, offset + 2).toString('hex');
      offset += 2;
      const tlvLength = buf.slice(offset, offset + 2).readUInt16BE(0);
      offset += 2;
      const tlvData = buf.slice(offset, offset + tlvLength);
      offset += tlvLength;

      let data = tlvData.toString('ascii') as string;
      if (section.description.includes("IP")) {
        data = num2dot(tlvData.readUInt32BE(0));
      }

      rows.push([
        chunkString(tlvData.toString('hex'), 2),
        data,
        tlvType + ':' + section.description,
      ]);
    } else if (section.isRepeat && section.repeatSpecs) {
      if (section.size !== -1) {
        repeat = value;
      }

      let specs : Spec[] = [];
      for (let i = 0; i < repeat; i++) {
        specs.push(...section.repeatSpecs);
      }

      const subrows : any[] = parseBuffer(buf.slice(offset), specs, repeat);
      rows.push(...subrows);
    } else {
      rows.push([chunkString(bufStr, 2).join(' '), value, section.description]);
    }
  }

  return rows;
}

function printBuffer(buf : Buffer, spec : Spec[]) {
  const rows = parseBuffer(buf, spec);
  console.log((new Table(rows)).toString());
}

function bufferFromWebText(webtext : string) : Buffer {
  return Buffer.from(webtext.replace(/\s/g, ''), 'hex');
}

const SNAC_01_0F = [
  byte("FLAP Header"),
  byte("Channel"),
  word("Sequence ID"),
  word("Payload Length"),
  word("SNAC Family"),
  word("SNAC Subtype"),
  word("SNAC Flags"),
  dword("SNAC Request-ID"),
  param(8, "UIN String Length"),
  word("Warning Level"),
  word("Number of TLV in list"),

  tlv("User Class"),
  tlv("User Status"),
  tlv("External IP Address"),
  tlv("Client Idle Time"),
  tlv("Signon Time"),
  tlv("Unknown Value"),
  tlv("Member Since"),

  word("DC Info"),
  word("DC Info Length"),
  dword("DC Internal IP Address"),
  dword("DC TCP Port"),
  dword("DC Type"),
  word("DC Protocol Version"),
  dword("DC Auth Cookie"),
  dword("Web Front Port"),
  dword("Client Features"),
  dword("Last Info Update Time"),
  dword("Last EXT info update time"),
  dword("Last EXT status update time"),
];

const exampleWebText = ''+
`
2a 02 00 05 00 71 00 01
00 0f 00 00 00 00 00 00
03 34 30 30 00 00 00 08
00 01 00 01 80 00 06 00
04 00 00 04 01 00 0a 00
04 c0 a8 01 fe 00 0f 00
04 00 00 00 00 00 03 00
04 61 3f aa 53 00 1e 00
04 00 00 00 00 00 05 00
04 36 78 bf a0 00 0c 00
26 c0 a8 01 fe 00 00 16
44 04 00 00 00 04 00 00
00 00 00 00 00 00 00 00
00 03 00 00 00 00 00 00
00 00 00 00 00 00 00
`

if (require.main === module) {
  printBuffer(bufferFromWebText(exampleWebText), SNAC_01_0F);
}
