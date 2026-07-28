import { execFile } from "node:child_process";
import { promisify } from "util";

const execFileAsync = promisify(execFile);

export const handler = async event => {
  if (event.rawPath !== "/plago") {
    return {
      statusCode: 404,
      body: "Not Found",
    };
  }

  console.info(event);
  const queryParams = event.queryStringParameters || {};
  const args = [];
  for (const [key, value] of Object.entries(queryParams)) {
    args.push(`--${key}`, value ?? "");
  }

  const promise = execFileAsync("./plago", args);

  if (
    event.requestContext?.http?.method === "POST" &&
    event.headers?.["content-type"].startsWith("application/json") &&
    event.body !== undefined
  ) {
    let body = event.body ?? "";
    if (event.isBase64Encoded) {
      body = Buffer.from(event.body, "base64").toString("utf8");
    }

    try {
      JSON.parse(body);
    } catch (e) {
      return {
        statusCode: 400,
        headers: { "content-type": "application/json" },
        body: JSON.stringify({ error: "Invalid JSON", message: e.message }),
      };
    }
    promise.child.stdin.write(body);
    promise.child.stdin.end();
  }

  try {
    const { stdout, stderr } = await promise;
    console.info(stderr);
    console.debug(stdout);

    return {
      statusCode: 200,
      headers: { "content-type": "application/json" },
      body: stdout,
    };
  } catch (e) {
    console.error(e);
    return {
      statusCode: 500,
      headers: { "content-type": "application/json" },
      body: e.stderr,
    };
  }
};
