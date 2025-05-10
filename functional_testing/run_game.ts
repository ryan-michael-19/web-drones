import type {paths, components} from "./types";
import createClient, {type Middleware} from "openapi-fetch";

import {writeFileSync, readFileSync} from "fs";
import process from "process";

const SLEEP_TIME = 1000;

function distanceSquared (x: number, y: number): number {
    return (x**2) + (y**2);
}
function eps(a: number, b: number) {
    return Math.abs(a-b) < 1e-9;
}

async function RunGame(username: string, password: string) {
    const authMiddleware: Middleware = {
        async onRequest({ schemaPath, request }) {
            if (schemaPath === "/login" || schemaPath === "/newUser") {
                const authString = btoa(`${username}:${password}`)
                request.headers.set("Authorization", `Basic ${authString}`);
            } else {
                const cookieData = readFileSync(`./cookies/${username}-cookie`);
                // for all other paths, set Authorization header as expected
                request.headers.set("Cookie",`${cookieData}`);
            }
            return request;
        },
        async onResponse({ schemaPath, response, options }) {
            // save cookie to file
            if (schemaPath === "/login" || schemaPath === "/newUser") {
                const cookie = response.headers.getSetCookie()[0];
                // cookie will be undefined if there is an HTTP error
                if (cookie !== undefined) {
                    writeFileSync(`./cookies/${username}-cookie`, cookie);
                }
            }
            return undefined; // don't modify anything!!
        },
    };

    const CLIENT = createClient<paths>({
        baseUrl: "http://localhost",
        // baseUrl: "https://webdrones.net",
        headers: {
            "Content-Type" : "plain/text"
        }
    });
    CLIENT.use(authMiddleware);
    let bots:components["schemas"]["Bot"][] = [];
    await (async ()=>{
        console.log("NEW USER");
        const {data, error} = await CLIENT.POST("/newUser");
        await new Promise(r => setTimeout(r, SLEEP_TIME));
        if (error !== undefined) { // try to log in if there's a new user error
            console.log("Error making new user", error.toString())
            console.log("LOGGING IN");
            // const {data, error} = await CLIENT.POST("/login", {parseAs:"text"});
            await CLIENT.POST("/login", {parseAs:"text"});
            await new Promise(r => setTimeout(r, SLEEP_TIME));
            const res1 = await CLIENT.POST("/init");
            await new Promise(r => setTimeout(r, SLEEP_TIME));
            console.log(res1.error)
            bots = res1.data.bots;
        } else { // no error
            bots = data.bots;
        }
    })();

    
    const res2 = await CLIENT.GET("/mines");
    await new Promise(r => setTimeout(r, SLEEP_TIME));
    const mines = res2.data;
    await Promise.all(bots.map(async bot => {
        const params = {path: {botId:bot.identifier}}
        // Mine for three scrap metal
        for (let i = 0; i < 3; i++) {
            const enumeratedMines = Array.from(mines.entries())
            const [closestMineIDX, closestMine] = enumeratedMines.reduce(
                (firstValue, secondValue, _) => {
                    const [firstIDX, firstMine] = firstValue;
                    const [secondIDX, secondMine] = secondValue;
                    const firstDistanceX =  firstMine.x - bot.coordinates.x;
                    const firstDistanceY =  firstMine.y - bot.coordinates.y;
                    const firstDistanceSquared = distanceSquared(firstDistanceX, firstDistanceY);

                    const secondDistanceX = secondMine.x - bot.coordinates.x;
                    const secondDistanceY = secondMine.y - bot.coordinates.y;
                    const secondDistanceSquared = distanceSquared(secondDistanceX, secondDistanceY);
                    const closerMine = firstDistanceSquared < secondDistanceSquared ?  firstMine : secondMine;
                    const closerMineIDX = firstDistanceSquared < secondDistanceSquared ?  firstIDX : secondIDX;
                    return [closerMineIDX, closerMine];
                }
            );
            // Delete mine from array after reduce so we don't send a second bot to it.
            mines.splice(closestMineIDX, 1)

            await CLIENT.POST("/bots/{botId}/move", {
                params: params,
                body: {
                    x: closestMine.x,
                    y: closestMine.y
                }
            });
            await new Promise(r => setTimeout(r, SLEEP_TIME));
            // TODO: Calculate time it takes bot to get to destination instead of looping
            // TODO: unlearn my functional programming brainrot
            const waitTillBotReachesDestination = async (w: ()=>Promise<boolean>) => {
                await new Promise(r => setTimeout(r, 1000));
                await w() ? null : await waitTillBotReachesDestination(w);
            }
            await waitTillBotReachesDestination(async () => {
                const [x,y] = [closestMine.x, closestMine.y];
                console.log(`${username}: Waiting for ${bot.identifier} to reach ${x},${y}`);
                const {data, error} = await CLIENT.GET("/bots/{botId}", {
                    params: {
                        path: {
                            botId: bot.identifier
                        }
                    }
                });
                await new Promise(r => setTimeout(r, SLEEP_TIME));
                return eps(data.coordinates.x, closestMine.x) && eps(data.coordinates.y, closestMine.y);
            });
            console.log(`${username}: ${bot.identifier} is mining at ${closestMine}`);
            await CLIENT.POST("/bots/{botId}/extract", {
                params: params
            });
            await new Promise(r => setTimeout(r, SLEEP_TIME));
        }
        // Make a new bot with the scrap metal
        await CLIENT.POST("/bots/{botId}/newBot", {params:params, body:{NewBotName:"New Bot"}});
    }));
    return await CLIENT.GET("/bots");
}

const CONFIG = {
    "single": [
        ["ryan", "pw"],
    ],
    "multi": [
        ["ellie", "pw"],
        ["jamie", "newpw"],
        ["sam", "anotherpw"],
    ]
};
// nO tOp LeVeL aWaIt
(async () => {
    // TODO: Run this in jest??
    Promise.all(
        CONFIG[process.argv[2]].map(async ([username, password])=>{
            const result = await RunGame(username, password);
            console.log(username);
            console.log(result);
            // Check that each bot made a new bot
            console.assert(result.data.length === 6);
        })
    );
})();
