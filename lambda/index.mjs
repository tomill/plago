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

  const queryParams = event.queryStringParameters || {};
  const args = [];
  for (const [key, value] of Object.entries(queryParams)) {
    args.push(`--${key}`, value ?? "");
  }
  console.info("plago", args.join(" "));

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

    JSON.parse(body); // to check json parsable
    promise.child.stdin.write(body);
    promise.child.stdin.end();
  }

  const { stdout, stderr } = await promise;
  console.debug(stdout);
  console.info(stderr);

  return {
    statusCode: 200,
    headers: { "content-type": "application/json" },
    body: stdout,
  };
};
