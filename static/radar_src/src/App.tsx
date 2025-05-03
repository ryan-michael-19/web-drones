import './App.css';

import { useState } from "react";
import { Login } from "./Login.tsx";
import { Radar } from "./Radar.tsx";

import { sendLoginRequest } from './WebClient.ts';

export enum LoginState {
  NoRequestSent = "NoRequestSent",
  SuccessfulLogin = "SuccessfulLogin",
  FailedLogin = "FailedLogin"
}
function App() {
  const [isLoggedIn, setIsLoggedIn] = useState<LoginState>(LoginState.NoRequestSent);
  return (
    <>
      <h1>Webdrones Radar!</h1>
      {
        isLoggedIn === LoginState.SuccessfulLogin ?
          <Radar/>
          :
          <>
            <Login 
              loginToggler={async (e: FormData) => setIsLoggedIn(
                await sendLoginRequest(
                  e.get("username") as string|null ?? "", 
                  e.get("password") as string|null ?? ""
                )
              )}
            />
            {
              isLoggedIn === LoginState.FailedLogin ? 
                <p>There was an issue logging in. Please try again.</p>
                :
                null
            }
          </>
      }
      <h2>PHOTOSENSITIVITY WARNING! Radar will flash and blink when you log in!</h2>
    </>
  )
}

export default App
