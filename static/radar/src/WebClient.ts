import createClient, { Middleware }  from "openapi-fetch";
import { paths } from "./types";

export const client = createClient<paths>({
  // baseUrl: "https://webdrones.net",
  baseUrl: "http://localhost",
  credentials: 'include',
  redirect: 'follow'
});

export async function sendLoginRequest(username: string, password: string) {
  // We should only ever be logging in once, so creating the middleware func
  // on every call is Probably Fine (famous last words)
  const middleware: Middleware = {
    async onRequest({ request, options }) {
      const authString = btoa(`${username}:${password}`)
      request.headers.set("Authorization", `Basic ${authString}`);
    }
  }
  client.use(middleware);
  const res = await client.POST("/login", {parseAs: 'text'});
  client.eject(middleware);
  if (res.response.status === 200) {
    return true;
  } else {
    return false;
  }
}