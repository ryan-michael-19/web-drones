import './App.css';

import { useState } from "react";
import { Login } from "./Login.tsx";
import { Radar } from "./Radar.tsx";

import { sendLoginRequest } from './WebClient.ts';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  return (
    <>
      <h1>Webdrones Radar!</h1>
      {
        isLoggedIn ? 
          <Radar/>
          :
          <Login 
            loginToggler={async (e: FormData) => setIsLoggedIn(
              await sendLoginRequest(
                e.get("username") as string|null ?? "", 
                e.get("password") as string|null ?? ""
              )
            )}
          />
      }
      <h2>PHOTOSENSITIVITY WARNING! Radar will flash and blink when you log in!</h2>
    </>
  )
}

export default App
