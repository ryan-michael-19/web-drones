import './App.css';

import { useState } from "react";
import { Login } from "./Login.tsx";
import { Radar } from "./Radar.tsx";
import { sendLoginRequest } from './WebClient.ts';

function App() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  return (
    <>
      <h1>Webdrones Radar!</h1>
      {
        isLoggedIn ? 
          <Radar/>
          :
          <Login 
            loginToggler={async () => setIsLoggedIn(await sendLoginRequest(username, password))}
            usernameSetter={e => setUsername(e.target.value)}
            passwordSetter={e => setPassword(e.target.value)}
          />
      }
      <h2>PHOTOSENSITIVITY WARNING! Radar will flash and blink when you log in!</h2>
    </>
  )
}

export default App
