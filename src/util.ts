import { FLAP } from './structures'
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

const FLAPSpec = [
  byte("FLAP Header"),
  byte("Channel"),
  word("Sequence ID"),
  word("Payload Length"),
  word("SNAC Family"),
  word("SNAC Subtype"),
  word("SNAC Flags"),
  dword("SNAC Request-ID"),
  repeat(16, "Number of rate classes", [
    word("Rate Class ID"),
    dword("Window Size"),
    dword("Clear Level"),
    dword("Alert Level"),
    dword("Limit Level"),
    dword("Disconnect Level"),
    dword("Current Level"),
    dword("Max Level"),
    dword("Last Time"),
    byte("Current State")
  ]),
  repeat(-1, "", [
    word("Rate Group ID"),
    repeat(16, "Number of pairs in group", [
      dword("Family/Subtype pair"),
    ]),
  ]),
];

function parseBuffer(buf : Buffer, spec : Spec[], repeatTimes = 0) {
  let offset = 0;
  let rows = [];
  let repeat = repeatTimes;

  for (let section of spec) {
    let value : number = 0;
    if (section.size === 8) {
      const bufStr = buf.slice(offset, offset + 1).toString('hex');
      value = buf.readInt8(offset);
      rows.push([chunkString(bufStr, 2).join(' '), value, section.description]);
      offset += 1;
    } else if (section.size === 16) {
      const bufStr = buf.slice(offset, offset + 2).toString('hex');
      value = buf.readUInt16BE(offset);
      rows.push([chunkString(bufStr, 2).join(' '), value, section.description]);
      offset += 2;
    } else if (section.size === 32) {
      const bufStr = buf.slice(offset, offset + 4).toString('hex');
      value = buf.readUInt32BE(offset);
      rows.push([chunkString(bufStr, 2).join(' '), value, section.description]);
      offset += 4;
    }

    if (section.isRepeat && section.repeatSpecs) {
      if (section.size !== -1) {
        repeat = value;
      }

      let specs : Spec[] = [];
      for (let i = 0; i < repeat; i++) {
        specs.push(...section.repeatSpecs);
      }

      const subrows : any[] = parseBuffer(buf.slice(offset), specs, repeat);
      rows.push(...subrows);
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

const exampleWebText = ''+
`
2a 02 00 03 00 37 00 01
00 07 00 00 00 00 00 00
00 01 00 01 00 00 00 50
00 00 09 c4 00 00 07 d0
00 00 05 dc 00 00 03 20
00 00 0d 48 00 00 17 70
00 00 00 00 00 00 01 00
01 00 00 00 00                           `

if (require.main === module) {
  printBuffer(bufferFromWebText(exampleWebText), FLAPSpec);
}
