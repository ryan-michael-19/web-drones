import createClient, { Middleware }  from "openapi-fetch";
import { paths } from "./types";
import {LoginState} from "./App.tsx";

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
    // options is a required part of the callback type despite not being used
    // @ts-expect-error
    async onRequest({ request, options }) {
      const authString = btoa(`${username}:${password}`)
      request.headers.set("Authorization", `Basic ${authString}`);
    }
  }
  client.use(middleware);
  const res = await client.POST("/login", {parseAs: 'text'});
  client.eject(middleware);
  if (res.response.status === 200) {
    return LoginState.SuccessfulLogin;
  } else {
    return LoginState.FailedLogin;
  }
}