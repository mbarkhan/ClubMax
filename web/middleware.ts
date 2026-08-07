import {NextRequest, NextResponse} from "next/server";

export function middleware(request: NextRequest) {
  const {pathname} = request.nextUrl;
  if (pathname === "/") return NextResponse.redirect(new URL("/fa", request.url));
  if (!pathname.startsWith("/fa") && !pathname.startsWith("/en")) {
    return NextResponse.redirect(new URL(`/fa${pathname}`, request.url));
  }
  return NextResponse.next();
}

export const config = {matcher: ["/((?!api|_next|.*\\..*).*)"]};
