import { useRef, useEffect } from "react";
import { client } from "./WebClient";
import type{ components } from "./types";

function getImageName(bot: components["schemas"]["Bot"]): string {
    const imageRoot = "/src/assets/"
    if (bot.status === "IDLE") {
        if (bot.inventory <= 0) {
            return imageRoot+"drone_wave.png";
        } else {
            return imageRoot+"drone_scrap.png";
        }
    } else { // Bot is moving (this condition will change in the future!)
        if (bot.inventory <= 0) {
            return imageRoot+"drone_walk.png";
        } else {
            return imageRoot+"drone_walk_scrap.png";
        }
    }
} 

function distanceSquared(p: components["schemas"]["Coordinates"], q: components["schemas"]["Coordinates"]): number {
    return Math.pow(q.x-p.x, 2) + Math.pow(q.y-p.y, 2)
}

function updateCoords(currentCoords:components["schemas"]["Coordinates"], alreadyDrawn: components["schemas"]["Coordinates"][]) {
    // TODO: Think this through. It's not working
    // TODO: There's gotta be a faster/cleaner way to do this
    // Shift each bot name down if it's too close to an already drawn bot name
    function update(coordNumber: number) {
        if (coordNumber < alreadyDrawn.length) {
            const dist = 4;
            if (distanceSquared(currentCoords, alreadyDrawn[coordNumber]) < dist*dist) {
                currentCoords.y = alreadyDrawn[coordNumber].y-dist;
                return update(0);
            }
            else {
                return update(coordNumber+1);
            }
        } else {
            return currentCoords;
        }
    }
    return update(0);
}

function draw(context: CanvasRenderingContext2D, canvas:HTMLCanvasElement, bots: components["schemas"]["Bot"][], mines: components["schemas"]["Coordinates"][]) {
    context.fillStyle = '#FFFFFF';
    context.clearRect(-canvas.width*0.5, -canvas.height*0.5, canvas.width, canvas.height);
    const mineXDistances = mines.map(
        mine => mine.x
    );
    // Get the width of the minefield, and scale it to the width of the canvas.
    const mineWidth = Math.max(...mineXDistances) - Math.min(...mineXDistances);
    // const mineYDistances = mines.map(
        // mine => Math.abs(Math.abs(mine.y))
    // );
    // const furthestY = Math.max(...mineYDistances);
    // const scale = canvas.height/furthestY;
    // const scale = 13;
    const mineYDistances = mines.map(
        mine => mine.y
    );
    const mineHeight = Math.max(...mineYDistances) - Math.min(...mineYDistances);
    let scale;
    if (canvas.height < canvas.width) {
        scale = canvas.height/mineHeight;
    } else { // canvas.height >= canvas.width
        scale = canvas.width/mineWidth;
    }
    // make a small margin
    scale /= 1.9;
    // const scale = mineWidth/canvas.width;
    const drawnBotCoords: components["schemas"]["Coordinates"][] = [];
    bots.forEach(bot => {
        // TODO: make this static?? 
        const botImage = new Image();
        botImage.src = getImageName(bot);
        botImage.onload = (e) =>  {
            context.drawImage(botImage, 
                bot.coordinates.x*scale-25,
                bot.coordinates.y*scale-25, 
                50, 50);
        }
        // TODO: Think through rendering names. It's currently not working.
        // Shift all coordinates downward to accomodate bot images
        // const textCoordsToDraw = updateCoords({...bot.coordinates}, drawnBotCoords);
        // context.fillText(bot.name, textCoordsToDraw.x*scale, textCoordsToDraw.y*scale+100);
        // // context.fillText(bot.name, coordsToDraw.x, coordsToDraw.y);
        // drawnBotCoords.push(textCoordsToDraw);
    });
    // TODO: change render for mines that are next to each other
    mines.forEach(mine => {
        // const nearbyMines = getNearbyMineCount(mine, mines);
        context.fillText("X", mine.x*scale, mine.y*scale);
    });
}
function adjustCanvas(canvas:HTMLCanvasElement) {
    canvas.height=window.innerHeight*0.7;
    canvas.width=window.innerWidth*0.7;

}

function adjustContext(canvas:HTMLCanvasElement, context:CanvasRenderingContext2D) {
    // Keep in mind this does not translate mouse coordinates!!!
    context.translate(
        canvas.width * 0.5,
        canvas.height * 0.5
    );
}

export function Radar() {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    useEffect(() => {
        const canvas = canvasRef.current;
        let interval;
        if (canvas !== null) {
            adjustCanvas(canvas);
            window.addEventListener('resize', () => adjustCanvas(canvas));
            const context = canvas.getContext('2d');
            if (context !== null){
                // because we are using setInterval we shouldn't need to use
                // requestAnimationFrame (I think)
                // TODO: Synchronize fetch requests with an animation loop so 
                // window resizes dynamically refit canvas data
                adjustContext(canvas,context);
                window.addEventListener('resize', () => adjustContext(canvas, context));
                interval = setInterval(async () => {
                    const bots = await client.GET("/bots");
                    const mines = await client.GET("/mines");
                    if (bots.data && mines.data) {
                        draw(context, canvas, bots.data, mines.data);
                    } else {
                        // TODO: Handle more elegantly
                        let errString;
                        if (!bots.data && mines.data) {
                            errString = "bots";
                        } else if (bots.data && !mines.data) {
                            errString = "mines";
                        } else if (!bots.data && !mines.data) {
                            errString = "bots and mines";
                        } else {
                            // We should never get here. but, yaknow.
                            throw new Error("Unspecifed fetch error.");
                        }
                        throw new Error(`Cannot find ${errString}`);
                    }
                }, 1000);
            } else {
                throw new Error("Canvas context is null");
            }
        } else {
            throw new Error("Canvas is null");
        }
        // We should never end up here, but better safe than sorry!
        return () => {
            if (interval !== undefined) { 
                clearInterval(interval);
            }
        }
    }, [draw]);
    // }, []);
  
    return (
        <>
            {/* <canvas ref={canvasRef} height={window.innerWidth*0.5} width={window.innerWidth*0.7}/> */}
            <canvas ref={canvasRef}></canvas>
        </>
    )
}