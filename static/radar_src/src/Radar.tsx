import { useRef, useEffect, useState } from "react";
import { client } from "./WebClient";
import type { components } from "./types";

import droneWaveImage from "./assets/drone_wave.png";
import droneScrapImage from "./assets/drone_scrap.png";
import droneWalkImage from "./assets/drone_walk.png";
import droneWalkWithScrapImage from "./assets/drone_walk_scrap.png";

function getImageName(bot: components["schemas"]["Bot"]): string {
    // const imageRoot = "/src/assets/"
    if (bot.status === "IDLE") {
        if (bot.inventory <= 0) {
            // return imageRoot+"drone_wave.png";
            return droneWaveImage;
        } else {
            // return imageRoot+"drone_scrap.png";
            return droneScrapImage;
        }
    } else { // Bot is moving (this condition will change in the future!)
        if (bot.inventory <= 0) {
            // return imageRoot+"drone_walk.png";
            return droneWalkImage;
        } else {
            // return imageRoot+"drone_walk_scrap.png";
            return droneWalkWithScrapImage;
        }
    }
} 

// function distanceSquared(p: components["schemas"]["Coordinates"], q: components["schemas"]["Coordinates"]): number {
    // return Math.pow(q.x-p.x, 2) + Math.pow(q.y-p.y, 2)
// }

function fitScaleToMines(mines: components["schemas"]["Coordinates"][], canvas: HTMLCanvasElement): number {
    const mineXDistances = mines.map(
        mine => mine.x
    );
    const mineWidth = Math.max(...mineXDistances) - Math.min(...mineXDistances);
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
    return scale;
}

function draw(context: CanvasRenderingContext2D, canvas:HTMLCanvasElement, bots: components["schemas"]["Bot"][], mines: components["schemas"]["Coordinates"][]) {
    context.fillStyle = '#FFFFFF';
    context.clearRect(-canvas.width*0.5, -canvas.height*0.5, canvas.width, canvas.height);
    setTimeout(() => {
        const scale = fitScaleToMines(mines, canvas);
        bots.forEach(bot => {
            // TODO: make this static?? 
            const botImage = new Image();
            botImage.src = getImageName(bot);
            botImage.onload = (_) =>  {
                context.drawImage(botImage, 
                    bot.coordinates.x*scale-25,
                    bot.coordinates.y*scale-25,
                    50, 50);
                context.fillText(bot.name, (bot.coordinates.x*scale), bot.coordinates.y*scale+35);
            }
        });
        // TODO: change render for mines that are next to each other
        mines.forEach(mine => {
            // const nearbyMines = getNearbyMineCount(mine, mines);
            context.fillText("X", mine.x*scale, mine.y*scale);
        });
    }, 300);
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

async function updateData(setBots: (b: components["schemas"]["Bot"][]) => void, setMines: (m: components["schemas"]["Coordinates"][]) => void) {
    const b = await client.GET("/bots");
    await new Promise(r => setTimeout(r, 1100));
    const m = await client.GET("/mines");
    await new Promise(r => setTimeout(r, 1100));
    // TODO: Ensure react is only triggering one render here
    setBots(b.data ? b.data : []);
    setMines(m.data ? m.data : []);
}

export function Radar() {
    const canvasRef = useRef<HTMLCanvasElement>(null);
    const [bots, setBots] = useState<components["schemas"]["Bot"][]>([]);
    const [mines, setMines] = useState<components["schemas"]["Coordinates"][]>([]);
    useEffect(() => {
        // updateData(setBots, setMines);
        // this interval MUST be larger than any sleeping we do with the logic
        // inside the loop!!! Just trust me that bad stuff will happen otherwise
        const i = setInterval(() => updateData(setBots, setMines), 2500);
        return () => {
            clearInterval(i);
        }
    }, []);
    useEffect(() => {
        const canvas = canvasRef.current;
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
                function drawLoop() {
                    if (bots.length > 0 && mines.length > 0) {
                        draw(context!, canvas!, bots, mines);
                    } else {
                        // TODO: Handle more elegantly
                        let errString;
                        if (bots.length <= 0 && mines.length > 0) {
                            errString = "bots";
                        } else if (bots.length > 0 && mines.length <= 0) {
                            errString = "mines";
                        } else if (bots.length <=0 && mines.length <= 0) {
                            errString = "bots and mines";
                        } else {
                            // We should never get here. but, yaknow.
                            throw new Error("Unspecifed fetch error.");
                        }
                        // throw new Error(`Cannot find ${errString}`);
                        console.log(`Cannot find ${errString}`);
                    }
                    // requestAnimationFrame(drawLoop);
                }
                drawLoop();
            } else {
                throw new Error("Canvas context is null");
            }
        } else {
            throw new Error("Canvas is null");
        }
    }, [bots, mines]);
    // }, []);
  
    return (
        <>
            <canvas ref={canvasRef}></canvas>
        </>
    )
}