export function Login ({loginToggler}: {loginToggler: (formData: FormData) => void}) {
    return (
        <form id="login_container" action={loginToggler}>
            <div className="login_input">
                <label htmlFor="username">Enter Username: </label>
                <input type="text" name="username"/>
            </div>
            <div className="login_input">
                <label htmlFor="password">Enter Password: </label>
                <input type="password" name="password"/>
            </div>
            <div className="login_input">
                <input className="submit_button" type="submit" value="Log In"/>
            </div>
        </form>
    );
}