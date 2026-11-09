import { NextRequest, NextResponse } from "next/server";

export const dynamic = "force-dynamic";

function getBackendOrigin(): string {
  const url =
    process.env.API_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    "http://127.0.0.1:8081";
  return url.replace(/\/+$/, "");
}

async function proxyRequest(request: NextRequest): Promise<NextResponse> {
  const backendOrigin = getBackendOrigin();
  const pathname = request.nextUrl.pathname;
  const search = request.nextUrl.search;
  const targetUrl = `${backendOrigin}${pathname}${search}`;

  const forwardHeaders = new Headers();
  request.headers.forEach((value, key) => {
    const lower = key.toLowerCase();
    if (lower !== "host" && lower !== "connection" && lower !== "content-length") {
      forwardHeaders.set(key, value);
    }
  });

  const clientIp =
    request.headers.get("x-forwarded-for") ||
    request.headers.get("x-real-ip");
  if (clientIp) {
    forwardHeaders.set("x-forwarded-for", clientIp);
  }

  const method = request.method;
  const hasBody = !["GET", "HEAD"].includes(method);
  let body: ArrayBuffer | undefined;

  if (hasBody) {
    try {
      body = await request.arrayBuffer();
    } catch {
      body = undefined;
    }
  }

  try {
    const upstream = await fetch(targetUrl, {
      method,
      headers: forwardHeaders,
      body,
      redirect: "manual",
    });

    const responseHeaders = new Headers();
    upstream.headers.forEach((value, key) => {
      if (key.toLowerCase() !== "set-cookie") {
        responseHeaders.set(key, value);
      }
    });

    const cookies = upstream.headers.getSetCookie
      ? upstream.headers.getSetCookie()
      : [upstream.headers.get("set-cookie")].filter(Boolean) as string[];

    const response = new NextResponse(upstream.body, {
      status: upstream.status,
      statusText: upstream.statusText,
      headers: responseHeaders,
    });

    for (const cookie of cookies) {
      if (cookie) {
        response.headers.append("set-cookie", cookie);
      }
    }

    return response;
  } catch (error) {
    return NextResponse.json(
      {
        code: "BACKEND_UNAVAILABLE",
        message: "Failed to connect to backend service",
        details: error instanceof Error ? error.message : String(error),
      },
      { status: 502 }
    );
  }
}

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PUT = proxyRequest;
export const PATCH = proxyRequest;
export const DELETE = proxyRequest;
export const OPTIONS = proxyRequest;
export const HEAD = proxyRequest;
