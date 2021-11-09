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
  dump? : boolean,
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

function dump() : Spec {
  return {size: -1, description: '', dump: true};
}

function parseBuffer(buf : Buffer, spec : Spec[], repeatTimes = 0) {
  let offset = 0;
  let rows = [];
  let repeat = repeatTimes;

  for (let section of spec) {
    let value : any = 0;
    let bufStr : string = '';

    if (section.dump) {
      rows.push({raw: logDataStream(buf.slice(offset))});
      break;
    }

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

  const lastRow = rows[rows.length - 1];

  if (!!lastRow.raw) {
    console.log((new Table(rows.slice(0, -1))).toString());
    console.log(lastRow.raw);
  } else {
    console.log((new Table(rows)).toString());
  }
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

const exSNAC_01_0F = ''+
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
`;

const SNAC_01_07 = [
  byte("FLAP Header"),
  byte("Channel"),
  word("Sequence ID"),
  word("Payload Length"),
  word("SNAC Family"),
  word("SNAC Subtype"),
  word("SNAC Flags"),
  dword("SNAC Request-ID"),

  repeat(16, "Number of Rate Classes", [
    word('Rate class ID'),
    dword('Window size'),
    dword('Clear level'),
    dword('Alert level'),
    dword('Limit level'),
    dword('Disconnect level'),
    dword('Current level'),
    dword('Max level'),
    dword('Last time'),
    byte('Current State'),
  ]),

  dump(),
];

const exSNAC_01_07 = ''+
`
2a 02 00 05 03 3b 00 01 00 07 00 00 00 00
00 00 00 05 00 01 00 00 00 50 00 00 09 c4 00 00
07 d0 00 00 05 dc 00 00 03 20 00 00 16 dc 00 00
17 70 00 00 00 00 00 00 02 00 00 00 50 00 00 0b
b8 00 00 07 d0 00 00 05 dc 00 00 03 e8 00 00 17
70 00 00 17 70 00 00 00 7b 00 00 03 00 00 00 1e
00 00 0e 74 00 00 0f a0 00 00 05 dc 00 00 03 e8
00 00 17 70 00 00 17 70 00 00 00 00 00 00 04 00
00 00 14 00 00 15 7c 00 00 14 b4 00 00 10 68 00
00 0b b8 00 00 17 70 00 00 1f 40 00 00 00 7b 00
00 05 00 00 00 0a 00 00 15 7c 00 00 14 b4 00 00
10 68 00 00 0b b8 00 00 17 70 00 00 1f 40 00 00
00 7b 00 00 01 00 91 00 01 00 01 00 01 00 02 00
01 00 03 00 01 00 04 00 01 00 05 00 01 00 06 00
01 00 07 00 01 00 08 00 01 00 09 00 01 00 0a 00
01 00 0b 00 01 00 0c 00 01 00 0d 00 01 00 0e 00
01 00 0f 00 01 00 10 00 01 00 11 00 01 00 12 00
01 00 13 00 01 00 14 00 01 00 15 00 01 00 16 00
01 00 17 00 01 00 18 00 01 00 19 00 01 00 1a 00
01 00 1b 00 01 00 1c 00 01 00 1d 00 01 00 1e 00
01 00 1f 00 01 00 20 00 01 00 21 00 02 00 01 00
02 00 02 00 02 00 03 00 02 00 04 00 02 00 06 00
02 00 07 00 02 00 08 00 02 00 0a 00 02 00 0c 00
02 00 0d 00 02 00 0e 00 02 00 0f 00 02 00 10 00
02 00 11 00 02 00 12 00 02 00 13 00 02 00 14 00
02 00 15 00 03 00 01 00 03 00 02 00 03 00 03 00
03 00 06 00 03 00 07 00 03 00 08 00 03 00 09 00
03 00 0a 00 03 00 0b 00 03 00 0c 00 04 00 01 00
04 00 02 00 04 00 03 00 04 00 04 00 04 00 05 00
04 00 07 00 04 00 08 00 04 00 09 00 04 00 0a 00
04 00 0b 00 04 00 0c 00 04 00 0d 00 04 00 0e 00
04 00 0f 00 04 00 10 00 04 00 11 00 04 00 12 00
04 00 13 00 04 00 14 00 06 00 01 00 06 00 02 00
06 00 03 00 08 00 01 00 08 00 02 00 09 00 01 00
09 00 02 00 09 00 03 00 09 00 04 00 09 00 09 00
09 00 0a 00 09 00 0b 00 0a 00 01 00 0a 00 02 00
0a 00 03 00 0b 00 01 00 0b 00 02 00 0b 00 03 00
0b 00 04 00 0c 00 01 00 0c 00 02 00 0c 00 03 00
13 00 01 00 13 00 02 00 13 00 03 00 13 00 04 00
13 00 05 00 13 00 06 00 13 00 07 00 13 00 08 00
13 00 09 00 13 00 0a 00 13 00 0b 00 13 00 0c 00
13 00 0d 00 13 00 0e 00 13 00 0f 00 13 00 10 00
13 00 11 00 13 00 12 00 13 00 13 00 13 00 14 00
13 00 15 00 13 00 16 00 13 00 17 00 13 00 18 00
13 00 19 00 13 00 1a 00 13 00 1b 00 13 00 1c 00
13 00 1d 00 13 00 1e 00 13 00 1f 00 13 00 20 00
13 00 21 00 13 00 22 00 13 00 23 00 13 00 24 00
13 00 25 00 13 00 26 00 13 00 27 00 13 00 28 00
15 00 01 00 15 00 02 00 15 00 03 00 02 00 06 00
03 00 04 00 03 00 05 00 09 00 05 00 09 00 06 00
09 00 07 00 09 00 08 00 03 00 02 00 02 00 05 00
04 00 06 00 04 00 02 00 02 00 09 00 02 00 0b 00
05 00 00                                       
`;

if (require.main === module) {
  printBuffer(bufferFromWebText(exSNAC_01_07), SNAC_01_07);
}
