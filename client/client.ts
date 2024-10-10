import type {paths, components} from "./types";
import createClient, {type Middleware} from "openapi-fetch";

import {writeFileSync, readFileSync} from "fs";

const authMiddleware: Middleware = {
    async onRequest({ schemaPath, request }) {
        if (schemaPath === "/login" || schemaPath === "/newUser") {
            const username = "big chungus";
            const password = "get rekt";
            const authString = btoa(`${username}:${password}`)
            request.headers.set("Authorization", `Basic ${authString}`);
        } else {
            const cookieData = readFileSync("./muh-cookie");
            // for all other paths, set Authorization header as expected
            request.headers.set("Cookie",`${cookieData}`);
        }
        return request;
    },
    async onResponse({ schemaPath, response, options }) {
        // save cookie to file
         if (schemaPath === "/login" || schemaPath === "/newUser") {
            const cookie = response.headers.getSetCookie()[0];
            console.log(cookie);
            writeFileSync("./muh-cookie", cookie);
        }
        return undefined; // don't modify anything!!
    },
};

const CLIENT = createClient<paths>({
    baseUrl: "http://localhost:8080"
});
CLIENT.use(authMiddleware);
// CLIENT.POST("/login");

function distanceSquared (x: number, y: number): number {
    return (x**2) + (y**2);
}
function eps(a: number, b: number) {
    return Math.abs(a-b) < 1e-9;
}

// nO tOp lEvEl aWaIt
(async () => {
    // await CLIENT.POST("/login", {
        // params: {
            // header: {
// 
            // }
        // }
    // })
    await CLIENT.POST("/newUser");
    const {data, error} = await CLIENT.POST("/init");
    const bots = data.bots;
    await Promise.all(bots.map(async bot => {
        const params = {path: {botId:bot.identifier}}
        // Mine for three scrap metal
        for (let i = 0; i < 3; i++) {
            const {data, error} = await CLIENT.GET("/mines");
            const mines = data;
            const closestMine = mines.reduce(
                (firstMine, secondMine) => {
                    const firstDistanceX =  firstMine.x - bot.coordinates.x;
                    const firstDistanceY =  firstMine.y - bot.coordinates.y;
                    const firstDistanceSquared = distanceSquared(firstDistanceX, firstDistanceY);

                    const secondDistanceX = secondMine.x - bot.coordinates.x;
                    const secondDistanceY = secondMine.y - bot.coordinates.y;
                    const secondDistanceSquared = distanceSquared(secondDistanceX, secondDistanceY);
                    const closerMine = firstDistanceSquared < secondDistanceSquared ? firstMine : secondMine;
                    return closerMine;
                }
            );
            await CLIENT.POST("/bots/{botId}/move", {
                params: params,
                body: {
                    x: closestMine.x,
                    y: closestMine.y
                }
            });
            // TODO: Calculate time it takes bot to get to destination instead of looping
            // TODO: unlearn my functional programming brainrot
            const waitTillBotReachesDestination = async (w: ()=>Promise<boolean>) => {
                await new Promise(r => setTimeout(r, 1000));
                await w() ? null : await waitTillBotReachesDestination(w);
            }
            await waitTillBotReachesDestination(async () => {
                const [x,y] = [closestMine.x, closestMine.y];
                console.log(`Waiting for ${bot.identifier} to reach ${x},${y}`);
                const {data, error} = await CLIENT.GET("/bots/{botId}", {
                    params: {
                        path: {
                            botId: bot.identifier
                        }
                    }
                });
                return eps(data.coordinates.x, closestMine.x) && eps(data.coordinates.y, closestMine.y);
            });
            console.log(`${bot.identifier} is mining at ${closestMine}`);
            await CLIENT.POST("/bots/{botId}/mine", {
                params: params
            });
        }
        // Make a new bot with the scrap metal
        await CLIENT.POST("/bots/{botId}/newBot", {params:params, body:{NewBotName:"New Bot"}});
    }));
})();