// HTTP Response structure from the debugger
interface HttpResponse {
  status: number;
  headers: { [key: string]: string[] };
  body: ArrayBuffer;
}

// Script context passed from debugger
interface ScriptContext {
  request?: any;
  response?: HttpResponse;
}

// Script result returned to debugger
interface ScriptResult {
  modified: boolean;
  modifiedResponse?: HttpResponse;
  logs: string[];
}

/**
 * Main plugin function called by Extism runtime
 * This is triggered for every HTTP response that matches the script's rules
 */
export function transform(): number {
  const logs: string[] = [];
  logs.push("Script started: Transform JSON Response");

  try {
    // Read input from Extism host
    const inputStr = Host.inputString();
    const ctx: ScriptContext = JSON.parse(inputStr);

    if (!ctx.response) {
      logs.push("No response in context, skipping");
      const result: ScriptResult = { modified: false, logs };
      Host.outputString(JSON.stringify(result));
      return 0;
    }

    const response = ctx.response;
    logs.push(`Response status: ${response.status}`);

    // Check if Content-Type is JSON
    const contentType = response.headers["Content-Type"] || response.headers["content-type"];
    const isJson = contentType && contentType.some(ct => ct.includes("application/json"));

    if (!isJson) {
      logs.push("Response is not JSON, skipping transformation");
      const result: ScriptResult = { modified: false, logs };
      Host.outputString(JSON.stringify(result));
      return 0;
    }

    // Parse response body
    const bodyBytes = new Uint8Array(response.body);
    const bodyStr = new TextDecoder().decode(bodyBytes);
    logs.push(`Original body length: ${bodyStr.length} bytes`);

    let bodyJson: any;
    try {
      bodyJson = JSON.parse(bodyStr);
    } catch (e) {
      logs.push(`Failed to parse JSON body: ${e}`);
      const result: ScriptResult = { modified: false, logs };
      Host.outputString(JSON.stringify(result));
      return 0;
    }

    // Add custom metadata to response
    bodyJson._metadata = {
      processedAt: new Date().toISOString(),
      processedBy: "network-debugger-script",
      version: "1.0.0",
    };

    logs.push("Added _metadata field to response");

    // Serialize modified body
    const modifiedBodyStr = JSON.stringify(bodyJson);
    const modifiedBodyBytes = new TextEncoder().encode(modifiedBodyStr);
    logs.push(`Modified body length: ${modifiedBodyStr.length} bytes`);

    // Return modified response
    const modifiedResponse: HttpResponse = {
      status: response.status,
      headers: response.headers,
      body: modifiedBodyBytes.buffer,
    };

    const result: ScriptResult = {
      modified: true,
      modifiedResponse,
      logs,
    };

    Host.outputString(JSON.stringify(result));
    return 0;
  } catch (error) {
    logs.push(`Error: ${error}`);
    const result: ScriptResult = { modified: false, logs };
    Host.outputString(JSON.stringify(result));
    return 1;
  }
}
