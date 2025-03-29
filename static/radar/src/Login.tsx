import type { ChangeEvent } from "react";
export function Login (
    { loginToggler, usernameSetter, passwordSetter }: {
    loginToggler: () => void,
    usernameSetter: (e: ChangeEvent<HTMLInputElement>) => void,
    passwordSetter: (e: ChangeEvent<HTMLInputElement>) => void
}) {
    return (
        <div id="login_container">
            <div className="login_input">
                <label htmlFor="username">Enter Username: </label>
                <input type="text" id="username" onChange={usernameSetter}/>
            </div>
            <div className="login_input">
                <label htmlFor="password">Enter Password: </label>
                <input type="password" id="password" onChange={passwordSetter}/>
            </div>
            <button onClick={loginToggler}>Log In</button>
        </div>
    );
}