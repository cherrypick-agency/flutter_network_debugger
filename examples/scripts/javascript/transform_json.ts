// Пример JavaScript (AssemblyScript) скрипта для трансформации JSON
// Компиляция: npm install && npm run asbuild
// Требует: @extism/assemblyscript-pdk

import { JSON } from "assemblyscript-json/assembly";
import { Host, Config } from "@extism/as-pdk";

interface HTTPRequest {
  method: string;
  url: string;
  headers: Map<string, string[]>;
  body: ArrayBuffer;
}

interface HTTPResponse {
  status: i32;
  headers: Map<string, string[]>;
  body: ArrayBuffer;
}

// Main function вызывается Extism
export function process(): i32 {
  // Читаем input
  const input = Host.inputString();

  // Парсим JSON request
  const reqJson = <JSON.Obj>JSON.parse(input);
  const body = reqJson.getString("body")!.valueOf();

  Host.log("Processing JSON transformation");

  // Пытаемся распарсить body как JSON
  let bodyJson: JSON.Obj;
  try {
    bodyJson = <JSON.Obj>JSON.parse(body);
  } catch (e) {
    // Если body не JSON - пропускаем без изменений
    Host.outputString(input);
    return 0;
  }

  // Трансформируем JSON: добавляем metadata
  const metadata = new JSON.Obj();
  metadata.set("processedBy", "JavaScript Script");
  metadata.set("timestamp", new Date().toISOString());
  metadata.set("scriptVersion", "1.0.0");

  bodyJson.set("_metadata", metadata);

  // Формируем response
  const response = new JSON.Obj();
  response.set("method", reqJson.getString("method")!);
  response.set("url", reqJson.getString("url")!);
  response.set("headers", reqJson.getObj("headers")!);
  response.set("body", bodyJson.toString());

  Host.outputString(response.stringify());
  return 0;
}

// package.json должен содержать:
// {
//   "scripts": {
//     "asbuild": "asc assembly/index.ts --target release"
//   },
//   "devDependencies": {
//     "@extism/assemblyscript-pdk": "^1.0.0",
//     "assemblyscript": "^0.27.0"
//   }
// }
