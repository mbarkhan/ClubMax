import {NextRequest, NextResponse} from "next/server";

export function middleware(request: NextRequest) {
  const {pathname} = request.nextUrl;
  if (pathname.startsWith("/v1/") || pathname.startsWith("/media/")) return NextResponse.next();
  if (pathname === "/") return NextResponse.redirect(new URL("/fa", request.url));
  if (!pathname.startsWith("/fa") && !pathname.startsWith("/en")) {
    return NextResponse.redirect(new URL(`/fa${pathname}`, request.url));
  }
  return NextResponse.next();
}

export const config = {matcher: ["/((?!api|v1|media|_next|.*\\..*).*)"]};
